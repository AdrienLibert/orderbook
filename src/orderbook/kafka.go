package main

import (
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaClient struct {
	commonConfig   *sarama.Config
	consumerConfig *sarama.Config
	producerConfig *sarama.Config
	brokers        []string
	adminClient    *sarama.Consumer
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

	kc.adminClient = kc.GetConsumer()
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

func (kc *KafkaClient) ConsumerMessagesChan(topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := (*kc.adminClient).Partitions(topic)
	topics, err := (*kc.adminClient).Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("DEBUG: topics: ", topics)
	consumer, err := (*kc.adminClient).ConsumePartition(
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
