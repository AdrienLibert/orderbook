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

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp float64 `json:"timestamp"`
}

type Trade struct {
	LeftOrderId  string  `json:"left_order_id"`
	RightOrderId string  `json:"right_order_id"`
	Quantity     float64 `json:"quantity"`
	Action       string  `json:"action"`
	Status       string  `json:"status"`
}

type KafkaClient struct {
	consumer       *sarama.Consumer
	producer       *sarama.AsyncProducer
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
	return kc
}

func (kc *KafkaClient) GetConsumer() *sarama.Consumer {
	if kc.consumer != nil {
		return kc.consumer
	}
	consumer, err := sarama.NewConsumer(kc.brokers, kc.consumerConfig)
	kc.consumer = &consumer
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

func (kc *KafkaClient) GetProducer() *sarama.AsyncProducer {
	if kc.producer != nil {
		return kc.producer
	}
	producer, err := sarama.NewAsyncProducer(kc.brokers, kc.producerConfig)
	kc.producer = &producer
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
	// debuging here
	topics, err := master.Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("Topics: ", topics)
	consumer, err := master.ConsumePartition(
		topic,
		partitions[0], // only first partition for now
		sarama.OffsetOldest,
	)

	if err != nil {
		fmt.Printf("Topic %v Partitions %v", topic, partitions)
		panic(err)
	}
	fmt.Println("Start consuming topic: ", topic)

	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				errors <- consumerError
				fmt.Println("consumerError: ", consumerError.Err)
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

	_quoteTopic string
	_tradeTopic string
}

func cut(i int, xs *[]*Order) *Order {
	y := (*xs)[i]
	*xs = append((*xs)[:i], (*xs)[i+1:]...)
	return y
}

func (o *Orderbook) Process(inOrder *Order, producerChannel chan<- Trade) {
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
		fmt.Printf(
			"Executed trade: %s %f @ %f | Left Order ID: %s, Right Order ID: %s | "+
				"Left Order Quantity: %f, Right Order Quantity: %f\n",
			action, tradeQuantity, outOrder.Price, inOrder.OrderID, outOrder.OrderID,
			inOrder.Quantity, outOrder.Quantity,
		)

		inOrder.Quantity -= tradeQuantity
		outOrder.Quantity -= tradeQuantity

		// TODO: publish orderbook change

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

	producer := o.kafkaClient.GetProducer()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	produceChannel := make(chan Trade)
	go func(tradeMessage <-chan Trade) {
		for {
			select {
			case msg := <-tradeMessage:
				fmt.Println("New Message produced", msg)
				producedCount++
			case <-signals:
				fmt.Println("Interrupt is detected")
				(*producer).AsyncClose() // Trigger a shutdown of the producer.
				produceChannel <- Trade{}
			}
			// message := &sarama.ProducerMessage{Topic: o._tradeTopic, Value: sarama.StringEncoder(msg)}
		}
	}(produceChannel)

	consumeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				consumedCount++
				order, err := convertMessageToOrder(msg.Value)
				if err != nil {
					handleError(err)
				}
				o.Process(&order, produceChannel)
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("Received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("Interrupt is detected")
				consumeChannel <- struct{}{}
			}
		}
	}()
	<-consumeChannel
	fmt.Println("Closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}

func NewOrderBook(bid []*Order, ask []*Order, kafkaClient *KafkaClient) *Orderbook {
	o := new(Orderbook)
	o.bid = bid
	o.ask = ask
	o.kafkaClient = kafkaClient
	o._quoteTopic = "orders.topic"
	o._tradeTopic = "order.status.topic"
	return o
}

func handleError(err error) {
	fmt.Println("invalid message consummed:", err)
}

func convertMessageToOrder(messageValue []byte) (Order, error) {
	var order = &Order{}
	if err := json.Unmarshal(messageValue, order); err != nil {
		return Order{}, err
	}

	return *order, nil
}

// func convertTradeToMessage(Trade) []byte {
// 	return []byte{}
// }
//
// func produceMessages(producer sarama.AsyncProducer, signals chan os.Signal, topic string, msg string) {
//
// }

func main() {
	fmt.Println("Starting orderbook")
	o := NewOrderBook(
		[]*Order{},
		[]*Order{},
		NewKafkaClient(),
	)
	o.Start()
}
