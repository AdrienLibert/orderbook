package main

import (
	"encoding/json"
	"fmt"
	"math"
    "math/rand"
	"os"
	"os/signal"
	"time"
	"sync"
	"strconv"

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

var marketMax = 150.00

type Order struct {
	TraderID  string  `json:"trader_id"`
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp string `json:"timestamp"`
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
	trader_id			string
	limit_price			float64
	quantity			float64
	n_last_trades		int
	ema_param 			float64
	max_newton_iter		int
	max_newton_error	float64
	equilibrium_price	float64
	theta 				float64
	aggressiveness_buy 	float64
	aggressiveness_sell float64
	targetBuy  			float64
    targetSell 			float64

	kafkaClient 		*KafkaClient

	_quoteTopic      	string
	_tradeTopic      	string
	_pricePointTopic 	string
}

func NewTrader(trader_id string, limit_price float64, quantity float64, kafkaClient *KafkaClient) *Trader {
	t := new(Trader)
	t.trader_id = trader_id
	t.limit_price = limit_price
	t.quantity = quantity
	t.kafkaClient = kafkaClient

	t.n_last_trades = 5
	t.ema_param = 2 / float64(t.n_last_trades + 1)
	t.max_newton_iter = 10
	t.max_newton_error = 0.0001
	t.equilibrium_price = 0

	rand.Seed(time.Now().UnixNano())
    randomNumber := rand.Float64()
	t.theta = -1.0 * (5.0 * randomNumber)
	t.aggressiveness_buy = -1.0 * (0.3 * randomNumber)
	t.aggressiveness_sell = -1.0 * (0.3 * randomNumber)

	t.equilibrium_price = 0
	t._quoteTopic = "orders.topic"
	t._tradeTopic = "order.status.topic"
	t._pricePointTopic = "order.last_price.topic"
	return t
}

func (t *Trader) newtonMethod(theta_est float64) float64 {
	var rightSide float64
	var max_ float64
	if t.quantity > 0 {
		max_ = 0
		rightSide = (t.theta * (t.limit_price - t.equilibrium_price)) / (math.Exp(t.theta) - 1)
	} else {
		max_ = 1000 
		rightSide = (t.theta * (t.equilibrium_price - t.limit_price)) / (math.Exp(t.theta) - 1)
	}
	for i := 0; i < t.max_newton_iter; i++ {
		expTheta := math.Exp(theta_est)
		expThetaMinusOne := expTheta - 1
		theta_est = theta_est - (theta_est*(t.limit_price-t.equilibrium_price))/expThetaMinusOne
		if math.Abs(theta_est-rightSide) < t.max_newton_error {
			break
		}
		dfTheta := (max_-t.equilibrium_price)/expThetaMinusOne - (expTheta*(max_-t.equilibrium_price)*theta_est)/(math.Pow(expThetaMinusOne, 2))
		theta_est -= rightSide / dfTheta
	}
	if theta_est == 0 {
		return 0.000001
	}
	return theta_est
}

func (t *Trader) calculateTarget(limit_price float64, equilibrium_price float64, aggressiveness, theta float64, isBuy bool) float64 {
	var factor float64
	if limit_price < equilibrium_price {
		if aggressiveness >= 0 {
			return limit_price
		} else {
			factor = (math.Exp(-aggressiveness*theta) - 1) / (math.Exp(theta) - 1)
			if isBuy {
				return limit_price * (1 - factor)
			} else {
				return limit_price + (marketMax-limit_price)*factor
			}
		}
	} else {
		if aggressiveness >= 0 {
			factor = (math.Exp(aggressiveness*theta) - 1) / (math.Exp(theta) - 1)
			if isBuy {
				return equilibrium_price + (limit_price-equilibrium_price)*factor
			} else {
				return limit_price + (equilibrium_price-limit_price)*(1-factor)
			}
		} else {
			thetaEst := t.newtonMethod(t.theta)
			factor = (math.Exp(-aggressiveness*thetaEst) - 1) / (math.Exp(thetaEst) - 1)
			if isBuy {
				return equilibrium_price * (1 - factor)
			} else {
				return equilibrium_price + (marketMax-equilibrium_price)*factor
			}
		}
	}
}

func (t *Trader) updateTargetPrices() {
	if t.equilibrium_price == 0 {
		t.equilibrium_price = t.limit_price
	} else {
		t.equilibrium_price = t.ema_param*t.limit_price + (1-t.ema_param)*t.equilibrium_price
	}
	t.targetBuy = t.calculateTarget(t.limit_price, t.equilibrium_price, t.aggressiveness_buy, t.theta, true)
	t.targetSell = t.calculateTarget(t.limit_price, t.equilibrium_price, t.aggressiveness_sell, t.theta, false)
}

func (t *Trader) Trade(orderListChannel chan<- Order) {
	t.updateTargetPrices()
	var order_type string
	var target float64
	if t.quantity > 0 {
		order_type = "buy"
		target = t.targetBuy
	} else {
		order_type = "sell"
		target = t.targetSell
	}
	orderListChannel <- publishOrder(t.trader_id, t.quantity, target,order_type)
}


func (t *Trader) Start() {
	orderProducer := t.kafkaClient.GetProducer()

	master := t.kafkaClient.GetConsumer()
	consumer, errors := t.kafkaClient.Assign(*master, t._tradeTopic)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	orderListChannel := make(chan Order)
	t.Trade(orderListChannel)
	go func(orderMessage <-chan Order) {
		for {
			select {
			case msg := <-orderMessage:
				producerMessage := sarama.ProducerMessage{Topic: t._quoteTopic, Value: sarama.StringEncoder(convertOrderToMessage(msg))}
				par, off, err := (*orderProducer).SendMessage(&producerMessage)
				if err != nil {
					fmt.Printf("ERROR: producing order in partition %d, offset %d: %s", par, off, err)
				} else {
					fmt.Println("INFO: produced order:", producerMessage)
					producedCount++
				}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... closing order producer...")
				(*orderProducer).Close()
				orderListChannel <- Order{}
			}
		}
	}(orderListChannel)
	consumeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				trade, err := convertMessageToTrade(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					if (trade.OrderId == t.trader_id) && (trade.Status != "closed") {
						consumedCount++
					}
				}
				
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... Closing trade consummer...")
				consumeChannel <- struct{}{}
			}
		}
	}()
	<-consumeChannel
	fmt.Println("INFO: closing... processed", consumedCount, "messages and produced", producedCount, "messages")
	}

func publishOrder(trader_id string, quantity float64, target float64, order_type string) Order {

	newUUID := uuid.New()
	order := Order{
		TraderID: trader_id,
		OrderID:  newUUID.String(),
		OrderType: order_type,
		Price:    target,
		Quantity: quantity,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
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
	var j float64 = -1
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			j = j * -1
			rand.Seed(time.Now().UnixNano())
			t := NewTrader(
				strconv.Itoa(i),
				float64(rand.Intn(21)+90),
				float64(j) * float64(rand.Intn(21)),
				NewKafkaClient(),
			)
			t.Start()
		}(i)
	}

	wg.Wait()
}