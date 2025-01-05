package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/shopspring/decimal"

	"github.com/IBM/sarama"
)

type Order struct {
	orderId   string
	orderType string
	price     float64
	quantity  int
	timestamp int
}

type OrderRequest struct {
	order_id   string
	order_type string
	price      float64
	quantity   int
	timestamp  int
}

type KafkaClient struct {
	config   *sarama.Config
	consumer *sarama.Consumer
	brokers  []string
}

func NewKafkaClient() *KafkaClient {
	kc := new(KafkaClient)
	kc.brokers = []string{"localhost:9094"}
	kc.config = sarama.NewConfig()
	kc.config.ClientID = "go-orderbook-consumer"
	kc.config.Consumer.Return.Errors = true
	kc.config.Net.SASL.Enable = false
	kc.config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	kc.config.Consumer.Offsets.Initial = sarama.OffsetNewest
	return kc
}

func (kc *KafkaClient) GetConsumer() *sarama.Consumer {
	if kc.consumer != nil {
		return kc.consumer
	}
	consumer, err := sarama.NewConsumer(kc.brokers, kc.config)
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

func (kc *KafkaClient) Assign(master sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := master.Partitions(topic)
	// debuging here
	topics, err := master.Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("Partitions: ", topics)
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
				fmt.Println("Got message on topic ", topic, msg.Value)
			}
		}
	}(topic, consumer)

	return consumers, errors
}

type Orderbook struct {
	bid         map[decimal.Decimal][]Order
	ask         map[decimal.Decimal][]Order
	kafkaClient *KafkaClient

	_quoteTopic string
}

func (o Orderbook) Process(order *sarama.ConsumerMessage) {
	// in_order := json.Unmarshal([]byte(order.Value))
}

func (o Orderbook) Start() {
	master := o.kafkaClient.GetConsumer()
	consumer, errors := o.kafkaClient.Assign(*master, o._quoteTopic)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	msgCount := 0
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				// fmt.Println("Received messages:", string(msg.Key), string(msg.Value))
				var order, _ = convertMessageToOrderType(msg.Value)
				fmt.Println(order.orderId, order.orderType, string(msg.Value))
			case consumerError := <-errors:
				msgCount++
				fmt.Println("Received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
	fmt.Println("Closing... processed", msgCount, "messages")
}

func NewOrderBook(bid map[decimal.Decimal][]Order, ask map[decimal.Decimal][]Order, kafkaClient *KafkaClient) *Orderbook {
	o := new(Orderbook)
	o.bid = bid
	o.ask = ask
	o.kafkaClient = kafkaClient
	o._quoteTopic = "orders.topic"
	return o
}

func convertMessageToOrderType(messageValue []byte) (Order, error) {
	var request OrderRequest
	if er := json.Unmarshal(messageValue, &request); er != nil {
		return Order{}, er
	}

	order := Order{
		orderId:   request.order_id,
		orderType: request.order_type,
		price:     request.price,
		quantity:  request.quantity,
		timestamp: request.timestamp,
	}

	return order, nil
}

func main() {
	fmt.Println("Starting orderbook")
	o := NewOrderBook(
		make(map[decimal.Decimal][]Order),
		make(map[decimal.Decimal][]Order),
		NewKafkaClient(),
	)
	o.Start()
}
