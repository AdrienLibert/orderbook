package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type KafkaClient struct {
	commonConfig    *sarama.Config
	consumerConfig  *sarama.Config
	producerConfig  *sarama.Config
	brokers         []string
	tradeTopic      string
	quoteTopic      string
	pricePointTopic string
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

	kc.pricePointTopic = "order.last_price.topic"
	kc.tradeTopic = "trades.topic"
	kc.quoteTopic = "orders.topic"
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

func convertMessageToPricePoint(value []byte) (PricePoint, error) {
	var pp PricePoint
	err := json.Unmarshal(value, &pp)
	return pp, err
}
