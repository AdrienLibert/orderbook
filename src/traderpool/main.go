package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var marketMax = 120.00

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp float64 `json:"timestamp"`
}

type Trade struct {
	OrderId  string  `json:"order_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Action   string  `json:"action"`
	Status   string  `json:"status"`
}

type PricePoint struct {
	Price float64 `json:"price"`
}

type KafkaClient struct {
	commonConfig   *sarama.Config
	consumerConfig *sarama.Config
	producerConfig *sarama.Config
	brokers        []string
}

func NewKafkaClient() *KafkaClient {
	kc := new(KafkaClient)
	kc.brokers = []string{getenv("OB__KAFKA__BOOTSTRAP_SERVERS", "localhost:9094")}
	kc.commonConfig = sarama.NewConfig()
	kc.commonConfig.ClientID = "go-traderpool-consumer"
	kc.commonConfig.Net.SASL.Enable = false
	if getenv("OB__KAFKA__SECURITY_PROTOCOL", "PLAINTEXT") == "PLAINTEXT" {
		kc.commonConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	kc.consumerConfig = sarama.NewConfig()
	kc.consumerConfig.Consumer.Return.Errors = true
	kc.consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	kc.producerConfig = sarama.NewConfig()
	kc.producerConfig.Producer.Retry.Max = 5
	kc.producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	kc.producerConfig.Producer.Idempotent = true
	kc.producerConfig.Net.MaxOpenRequests = 1
	kc.producerConfig.Producer.Return.Successes = true
	return kc
}

func (kc *KafkaClient) GetConsumer() *sarama.Consumer {
	consumer, err := sarama.NewConsumer(kc.brokers, kc.consumerConfig)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	return &consumer
}

func (kc *KafkaClient) GetProducer() *sarama.SyncProducer {
	producer, err := sarama.NewSyncProducer(kc.brokers, kc.producerConfig)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	return &producer
}

func (kc *KafkaClient) Assign(master sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := master.Partitions(topic)
	topics, err := master.Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("DEBUG: topics: ", topics)
	consumer, err := master.ConsumePartition(
		topic,
		partitions[0], // TODO: only first partition for now
		sarama.OffsetOldest,
	)

	if err != nil {
		fmt.Printf("ERROR: topic %v partitions %v", topic, partitions)
		panic(err)
	}
	fmt.Println("INFO: start consuming topic: ", topic)

	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				errors <- consumerError
				fmt.Println("ERROR: not able to consume: ", consumerError.Err)
			case msg := <-consumer.Messages():
				consumers <- msg
			}
		}
	}(topic, consumer)

	return consumers, errors
}

type Trader struct {
	traderId           string
	limitPrice         float64
	quantity           float64
	nLastTrades        int
	emaParam           float64
	maxNewtonIter      int
	maxNewtonError     float64
	equilibriumPrice   float64
	theta              float64
	aggressivenessBuy  float64
	aggressivenessSell float64
	targetBuy          float64
	targetSell         float64

	kafkaClient *KafkaClient

	_quoteTopic      string
	_tradeTopic      string
	_pricePointTopic string
}

func NewTrader(traderId string, limitPrice float64, quantity float64, kafkaClient *KafkaClient) *Trader {
	t := new(Trader)
	t.traderId = traderId
	t.limitPrice = limitPrice
	t.quantity = quantity
	t.kafkaClient = kafkaClient

	t.nLastTrades = 5
	t.emaParam = 2 / float64(t.nLastTrades+1)
	t.maxNewtonIter = 10
	t.maxNewtonError = 0.0001

	t.equilibriumPrice = limitPrice

	randomNumber := rand.Float64()
	t.theta = -2.5 + 5.0*randomNumber
	t.aggressivenessBuy = -0.15 + 0.3*randomNumber
	t.aggressivenessSell = -0.15 + 0.3*randomNumber

	fmt.Printf("Trader %s -> Limit Price: %.2f, Agg Buy: %.3f, Agg Sell: %.3f\n",
		t.traderId, t.limitPrice, t.aggressivenessBuy, t.aggressivenessSell)

	t._quoteTopic = "orders.topic"
	t._tradeTopic = "trades.topic"
	t._pricePointTopic = "order.last_price.topic"
	return t
}

func (t *Trader) computeThetaEstimate() float64 {
	if math.Abs(t.theta) < 1e-8 {
		return 0.000001
	}

	var rightSide float64
	if t.quantity > 0 {
		rightSide = (t.theta * (t.limitPrice - t.equilibriumPrice)) / (math.Exp(t.theta) - 1)
	} else {
		rightSide = (t.theta * (t.equilibriumPrice - t.limitPrice)) / (math.Exp(t.theta) - 1)
	}

	rightSide = math.Max(-2.0, math.Min(2.0, rightSide))

	return rightSide
}

func (t *Trader) calculateTarget(limitPrice float64, equilibriumPrice float64, aggressiveness, theta float64, isBuy bool) float64 {
	var factor float64

	if limitPrice < equilibriumPrice {
		if aggressiveness >= 0 {
			return limitPrice
		} else {
			factor = (math.Exp(math.Abs(aggressiveness)*theta) - 1) / (math.Exp(theta) - 1 + 1e-8)
			factor = math.Max(0.2, math.Min(0.8, factor))

			if isBuy {
				return math.Max(80.0, limitPrice*(1-factor))
			} else {
				return limitPrice + (marketMax-limitPrice)*factor
			}
		}
	} else {
		if aggressiveness >= 0 {
			factor = (math.Exp(aggressiveness*theta) - 1) / (math.Exp(theta) - 1 + 1e-8)
			factor = math.Max(0.2, math.Min(0.8, factor))

			if isBuy {
				return equilibriumPrice + (limitPrice-equilibriumPrice)*factor
			} else {
				return math.Max(limitPrice, limitPrice+(equilibriumPrice-limitPrice)*(1-factor))
			}
		} else {
			thetaEst := math.Max(-2.0, math.Min(2.0, t.computeThetaEstimate()))
			factor = (math.Exp(math.Abs(aggressiveness)*thetaEst) - 1) / (math.Exp(thetaEst) - 1 + 1e-8)
			factor = math.Max(0.2, math.Min(0.8, factor))

			if isBuy {
				return math.Max(80.0, equilibriumPrice*(1-factor))
			} else {
				return equilibriumPrice + (marketMax-equilibriumPrice)*factor
			}
		}
	}
}

func (t *Trader) updateTargetPrices() {
	if t.equilibriumPrice == 0 {
		t.equilibriumPrice = t.limitPrice
	} else {
		t.equilibriumPrice = t.emaParam*t.limitPrice + (1-t.emaParam)*t.equilibriumPrice
	}
	t.targetBuy = t.calculateTarget(t.limitPrice, t.equilibriumPrice, t.aggressivenessBuy, t.theta, true)
	t.targetSell = t.calculateTarget(t.limitPrice, t.equilibriumPrice, t.aggressivenessSell, t.theta, false)
}

func (t *Trader) Trade(orderListChannel chan<- Order) Order {
	t.updateTargetPrices()
	var orderType string
	var target float64

	if t.quantity > 0 {
		orderType = "buy"
		target = t.targetBuy
	} else {
		orderType = "sell"
		target = t.targetSell
	}

	order := publishOrder(t.traderId, t.quantity, target, orderType)
	orderListChannel <- order
	return order
}

func (t *Trader) Start() {
	orderProducer := t.kafkaClient.GetProducer()
	if orderProducer == nil {
		fmt.Println("ERROR: Kafka producer is nil! Exiting.")
		return
	}

	master := t.kafkaClient.GetConsumer()
	consumer, errors := t.kafkaClient.Assign(*master, t._tradeTopic)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	orderListChannel := make(chan Order, 10)

	go func(orderMessage <-chan Order) {
		for msg := range orderMessage {
			producerMessage := sarama.ProducerMessage{
				Topic: t._quoteTopic,
				Value: sarama.StringEncoder(convertOrderToMessage(msg)),
			}
			par, off, err := (*orderProducer).SendMessage(&producerMessage)
			if err != nil {
				fmt.Printf("ERROR: producing order in partition %d, offset %d: %s\n", par, off, err)
			} else {
				fmt.Println("INFO: produced order:", msg)
				producedCount++
			}
		}
	}(orderListChannel)

	go func() {
		for {
			select {
			case <-signals:
				fmt.Println("INFO: Interrupt detected... Stopping traders.")
				close(orderListChannel)
				return
			default:
				t.Trade(orderListChannel)
				time.Sleep(time.Second * 2)
			}
		}
	}()

	consumeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				trade, err := convertMessageToTrade(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					if (trade.OrderId == t.traderId) && (trade.Status != "closed") {
						consumedCount++
					}
				}
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received consumerError:", consumerError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("INFO: Interrupt detected... Closing trade consumer...")
				consumeChannel <- struct{}{}
			}
		}
	}()

	<-consumeChannel
	fmt.Println("INFO: Closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}

func publishOrder(traderId string, quantity float64, target float64, orderType string) Order {

	newUUID := uuid.New()
	order := Order{
		OrderID:   newUUID.String(),
		OrderType: orderType,
		Price:     target,
		Quantity:  quantity,
		Timestamp: float64(time.Now().Unix()),
	}
	return order
}

func handleError(err error) {
	fmt.Println("ERROR: invalid message consummed:", err)
}

func convertOrderToMessage(order Order) []byte {
	message, err := json.Marshal(order)
	if err != nil {
		fmt.Println("ERROR: invalid order being converted to message:", err)
	}
	return message
}

func convertMessageToTrade(messageValue []byte) (Trade, error) {
	var trade = &Trade{}
	if err := json.Unmarshal(messageValue, trade); err != nil {
		return Trade{}, err
	}

	return *trade, nil
}

func main() {
	fmt.Println("INFO: starting traders")

	rand.Seed(time.Now().UnixNano())

	var j float64 = -1
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			j = j * -1
			t := NewTrader(
				strconv.Itoa(i),
				float64(rand.Intn(105-95+1)+95),
				float64(j)*float64(rand.Intn(11)+5),
				NewKafkaClient(),
			)
			t.Start()
		}(i)
	}

	wg.Wait()
}
