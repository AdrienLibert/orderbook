package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"

	"github.com/IBM/sarama"
)

// utils

func cut(i int, xs *[]*Order) *Order {
	y := (*xs)[i]
	*xs = append((*xs)[:i], (*xs)[i+1:]...)
	return y
}

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// TODO: move to models file
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
	kc.commonConfig.ClientID = "go-orderbook-consumer"
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

type Orderbook struct {
	bid         []*Order
	ask         []*Order
	kafkaClient *KafkaClient

	_quoteTopic      string
	_tradeTopic      string
	_pricePointTopic string
}

func NewOrderBook(bid []*Order, ask []*Order, kafkaClient *KafkaClient) *Orderbook {
	o := new(Orderbook)
	o.bid = bid
	o.ask = ask
	o.kafkaClient = kafkaClient
	o._quoteTopic = "orders.topic"
	o._tradeTopic = "order.status.topic"
	o._pricePointTopic = "order.last_price.topic"
	return o
}

func (o *Orderbook) Process(inOrder *Order, producerChannel chan<- Trade, pricePointChannel chan<- PricePoint) {
	var in, out *[]*Order
	var action string
	var comparator func(x, y float64) bool

	if inOrder.Quantity > 0 {
		in = &o.bid
		out = &o.ask
		action = "buy"
		comparator = func(x, y float64) bool { return x >= y }
	} else {
		in = &o.ask
		out = &o.bid
		action = "sell"
		inOrder.Quantity = -inOrder.Quantity
		comparator = func(x, y float64) bool { return x <= y }
	}

	for inOrder.Quantity > 0 && len(*out) > 0 && comparator(inOrder.Price, (*out)[0].Price) {
		outOrder := cut(0, out)
		fmt.Println(action, inOrder, outOrder)

		tradeQuantity := math.Min(inOrder.Quantity, outOrder.Quantity)
		price := outOrder.Price
		fmt.Printf(
			"INFO: Executed trade: %s %f @ %f | Left Order ID: %s, Right Order ID: %s | "+
				"Left Order Quantity: %f, Right Order Quantity: %f\n",
			action, tradeQuantity, price, inOrder.OrderID, outOrder.OrderID,
			inOrder.Quantity, outOrder.Quantity,
		)

		inOrder.Quantity -= tradeQuantity
		outOrder.Quantity -= tradeQuantity

		producerChannel <- publishTrade(inOrder, tradeQuantity, price, action)
		producerChannel <- publishTrade(outOrder, tradeQuantity, price, action)
		pricePointChannel <- publishPricePoint(price)

		if outOrder.Quantity > 0 {
			*out = append(*out, outOrder)
			sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
		}
	}

	if inOrder.Quantity > 0 {
		*in = append(*in, inOrder)
		sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
	}
}

func (o *Orderbook) Start() {
	master := o.kafkaClient.GetConsumer()
	consumer, errors := o.kafkaClient.Assign(*master, o._quoteTopic)

	tradeProducer := o.kafkaClient.GetProducer()
	pricePointProducer := o.kafkaClient.GetProducer()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	traderChannel := make(chan Trade)
	go func(tradeMessage <-chan Trade) {
		for {
			select {
			case msg := <-tradeMessage:
				producerMessage := sarama.ProducerMessage{Topic: o._tradeTopic, Value: sarama.StringEncoder(convertTradeToMessage(msg))}
				par, off, err := (*tradeProducer).SendMessage(&producerMessage)
				if err != nil {
					fmt.Printf("ERROR: producing trade in partition %d, offset %d: %s", par, off, err)
				} else {
					fmt.Println("INFO: produced trade:", producerMessage)
					producedCount++
				}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... closing trade producer...")
				(*tradeProducer).Close()
				traderChannel <- Trade{}
			}
		}
	}(traderChannel)

	pricePointChannel := make(chan PricePoint)
	go func(pricePointMessage <-chan PricePoint) {
		for {
			select {
			case msg := <-pricePointMessage:
				producerMessage := sarama.ProducerMessage{Topic: o._pricePointTopic, Value: sarama.StringEncoder(convertPricePointToMessage(msg))}
				par, off, err := (*pricePointProducer).SendMessage(&producerMessage)
				if err != nil {
					fmt.Printf("ERROR: producing price point in partition %d, offset %d: %s", par, off, err)
				} else {
					fmt.Println("INFO: produced price point:", producerMessage)
					producedCount++
				}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... closing price point producer...")
				(*pricePointProducer).Close()
				pricePointChannel <- PricePoint{}
			}
		}
	}(pricePointChannel)

	consumeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				order, err := convertMessageToOrder(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					fmt.Println("DEBUG: received quote:", order)
					consumedCount++
				}
				o.Process(&order, traderChannel, pricePointChannel)
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... Closing quote consummer...")
				consumeChannel <- struct{}{}
			}
		}
	}()
	<-consumeChannel
	fmt.Println("INFO: closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}

func handleError(err error) {
	fmt.Println("ERROR: invalid message consummed:", err)
}

func convertMessageToOrder(messageValue []byte) (Order, error) {
	var order = &Order{}
	if err := json.Unmarshal(messageValue, order); err != nil {
		return Order{}, err
	}

	return *order, nil
}

func publishTrade(inOrder *Order, tradeQuantity float64, price float64, action string) Trade {
	var status string

	if inOrder.Quantity == 0 {
		status = "closed"
	} else {
		status = "partial"
	}

	trade := Trade{
		OrderId:  inOrder.OrderID,
		Quantity: tradeQuantity,
		Price:    price,
		Action:   action,
		Status:   status,
	}

	return trade
}

func publishPricePoint(price float64) PricePoint {
	return PricePoint{
		Price: price,
	}
}

func convertTradeToMessage(trade Trade) []byte {
	message, err := json.Marshal(trade)
	if err != nil {
		fmt.Println("ERROR: invalid trade being converted to message:", err)
	}
	return message
}

func convertPricePointToMessage(princePoint PricePoint) []byte {
	message, err := json.Marshal(princePoint)
	if err != nil {
		fmt.Println("ERROR: invalid price point being converted to message:", err)
	}
	return message
}

func main() {
	fmt.Println("INFO: tarting orderbook")
	o := NewOrderBook(
		[]*Order{},
		[]*Order{},
		NewKafkaClient(),
	)
	o.Start()
}
