package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/shopspring/decimal"

	"github.com/IBM/sarama"
)

type Order struct {
}

type KafkaClient struct {
	config   *sarama.Config
	consumer *sarama.Consumer
	brokers  []string
}

func NewKafkaClient() *KafkaClient {
	kc := new(KafkaClient)
	kc.config = sarama.NewConfig()
	kc.config.ClientID = "go-orderbook-consumer"
	kc.config.Consumer.Return.Errors = true
	kc.brokers = []string{"localhost:9094"}
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
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()
	return &consumer
}

func (kc *KafkaClient) assign(master sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := master.Partitions(topic)
	// debuging here
	topics, err := master.Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("Partitions: ", topics)
	// debuging here, somehow cannot connect to local kafka ?
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

func SimpleOrderBook(bid map[decimal.Decimal][]Order, ask map[decimal.Decimal][]Order, kafkaClient *KafkaClient) *Orderbook {
	o := new(Orderbook)
	o.bid = bid
	o.ask = ask
	o.kafkaClient = kafkaClient
	o._quoteTopic = "orders.topic"
	return o
}

func (o Orderbook) Start() {
	master := o.kafkaClient.GetConsumer()
	consumer, errors := o.kafkaClient.assign(*master, o._quoteTopic)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	msgCount := 0
	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				fmt.Println("Received messages", string(msg.Key), string(msg.Value))
			case consumerError := <-errors:
				msgCount++
				fmt.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
	fmt.Println("Processed", msgCount, "messages")
}

func main() {
	fmt.Println("Starting orderbook")
	o := SimpleOrderBook(
		make(map[decimal.Decimal][]Order),
		make(map[decimal.Decimal][]Order),
		NewKafkaClient(),
	)
	o.Start()
}
