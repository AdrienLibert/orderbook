package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

type MockConsumerChan chan *sarama.ConsumerMessage

func (m MockConsumerChan) SendOrder(order Order) {
	orderBytes, err := json.Marshal(order)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal order: %v", err))
	}
	m <- &sarama.ConsumerMessage{
		Topic:     "orders.topic",
		Partition: 0,
		Offset:    0,
		Value:     orderBytes,
	}
}

func (m MockConsumerChan) SendTrade(trade Trade) {
	tradeBytes, err := json.Marshal(trade)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal trade: %v", err))
	}
	m <- &sarama.ConsumerMessage{
		Topic:     "trades.topic", // Assuming a different topic
		Partition: 0,
		Offset:    0,
		Value:     tradeBytes,
	}
}

func (m MockConsumerChan) SendPricePoint(pp PricePoint) {
	ppBytes, err := json.Marshal(pp)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal order: %v", err))
	}
	m <- &sarama.ConsumerMessage{
		Topic:     "order.last_price.topic",
		Partition: 0,
		Offset:    0,
		Value:     ppBytes,
	}
}

// MockKafkaClient wraps KafkaClient with mock producer and consumer.
type MockKafkaClient struct {
	*KafkaClient
	MockProducer sarama.SyncProducer
	MockConsumer sarama.Consumer
	MockChan     MockConsumerChan
}

// NewMockKafkaClient creates a new mock Kafka client for testing.
func NewMockKafkaClient(t *testing.T) *MockKafkaClient {
	kc := NewKafkaClient()
	mockProducer := mocks.NewSyncProducer(t, kc.producerConfig)
	mockConsumer := mocks.NewConsumer(t, kc.consumerConfig)
	mockChan := make(MockConsumerChan, 100)

	// Configure mock consumer to return topics and partitions
	mockConsumer.SetTopicMetadata(map[string][]int32{
		kc.tradeTopic:      {0},
		kc.quoteTopic:      {0},
		kc.pricePointTopic: {0},
	})

	return &MockKafkaClient{
		KafkaClient:  kc,
		MockProducer: mockProducer,
		MockConsumer: mockConsumer,
		MockChan:     mockChan,
	}
}

// GetProducer returns the mock sync producer.
func (mkc *MockKafkaClient) GetProducer() *sarama.SyncProducer {
	return &mkc.MockProducer
}

// GetConsumer returns the mock consumer.
func (mkc *MockKafkaClient) GetConsumer() *sarama.Consumer {
	return &mkc.MockConsumer
}

// Assign sets up channels for consuming messages from a topic, using the mock channel.
func (mkc *MockKafkaClient) Assign(master sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage, 100)
	errors := make(chan *sarama.ConsumerError, 100)

	partitions, err := master.Partitions(topic)
	if err != nil {
		fmt.Printf("ERROR: failed to get partitions for topic %v: %v\n", topic, err)
		panic(err)
	}

	_, err = master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		fmt.Printf("ERROR: topic %v partitions %v: %v\n", topic, partitions, err)
		panic(err)
	}

	go func() {
		defer close(consumers)
		defer close(errors)
		for msg := range mkc.MockChan {
			if msg.Topic == topic {
				consumers <- msg
			}
		}
	}()

	return consumers, errors
}

func TestNewMockKafkaClient(t *testing.T) {
	client := NewMockKafkaClient(t)

	if client.KafkaClient == nil {
		t.Fatal("Expected non-nil KafkaClient")
	}
	if client.brokers[0] != "localhost:9094" {
		t.Errorf("Expected broker localhost:9094, got %v", client.brokers)
	}
	if client.tradeTopic != "trades.topic" || client.quoteTopic != "orders.topic" || client.pricePointTopic != "order.last_price.topic" {
		t.Errorf("Expected topics trades.topic, orders.topic, order.last_price.topic, got %v, %v, %v",
			client.tradeTopic, client.quoteTopic, client.pricePointTopic)
	}

	if client.MockProducer == nil {
		t.Fatal("Expected non-nil MockProducer")
	}
	if client.MockConsumer == nil {
		t.Fatal("Expected non-nil MockConsumer")
	}

	if client.MockChan == nil {
		t.Fatal("Expected non-nil MockChan")
	}

	topics, err := client.MockConsumer.Topics()
	if err != nil {
		t.Fatalf("Failed to get topics: %v", err)
	}
	expectedTopics := map[string]bool{
		"trades.topic":           true,
		"orders.topic":           true,
		"order.last_price.topic": true,
	}
	for _, topic := range topics {
		if !expectedTopics[topic] {
			t.Errorf("Unexpected topic %v", topic)
		}
	}
}

func TestGetProducer(t *testing.T) {
	client := NewMockKafkaClient(t)
	producer := client.GetProducer()
	if producer == nil {
		t.Fatal("Expected non-nil producer")
	}
	if *producer != client.MockProducer {
		t.Error("Expected producer to match MockProducer")
	}
}

func TestGetConsumer(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()
	if consumer == nil {
		t.Fatal("Expected non-nil consumer")
	}
	if *consumer != client.MockConsumer {
		t.Error("Expected consumer to match MockConsumer")
	}
}

func TestSendOrder(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()

	// Assert MockConsumer to *mocks.Consumer to access ExpectConsumePartition
	mockConsumer, ok := client.MockConsumer.(*mocks.Consumer)
	if !ok {
		t.Fatal("Expected MockConsumer to be *mocks.Consumer")
	}

	pc := mockConsumer.ExpectConsumePartition(client.quoteTopic, 0, sarama.OffsetOldest)
	defer pc.Close()

	consumers, errors := client.Assign(*consumer, client.quoteTopic)

	order := Order{
		OrderID:   uuid.New().String(),
		OrderType: "limit",
		Price:     50.5,
		Quantity:  100,
		Timestamp: float64(time.Now().Unix()),
	}
	client.MockChan.SendOrder(order)
	close(client.MockChan)

	select {
	case msg := <-consumers:
		if msg.Topic != client.quoteTopic {
			t.Errorf("Expected topic %v, got %v", client.quoteTopic, msg.Topic)
		}
		var received Order
		if err := json.Unmarshal(msg.Value, &received); err != nil {
			t.Fatalf("Failed to unmarshal consumed message: %v", err)
		}
		if received.OrderID != order.OrderID {
			t.Errorf("Expected OrderID %v, got %v", order.OrderID, received.OrderID)
		}
		if received.Price != order.Price {
			t.Errorf("Expected Price %v, got %v", order.Price, received.Price)
		}
	case errMsg := <-errors:
		t.Fatalf("Received unexpected error: %v", errMsg.Err)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

func TestSendTrade(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()

	mockConsumer, ok := client.MockConsumer.(*mocks.Consumer)
	if !ok {
		t.Fatal("Expected MockConsumer to be *mocks.Consumer")
	}

	pc := mockConsumer.ExpectConsumePartition(client.tradeTopic, 0, sarama.OffsetOldest)
	defer pc.Close()

	consumers, errors := client.Assign(*consumer, client.tradeTopic)

	trade := Trade{
		OrderId:  uuid.New().String(),
		Quantity: 100,
		Price:    50.5,
		Action:   "buy",
		Status:   "",
	}
	client.MockChan.SendTrade(trade)
	close(client.MockChan)

	select {
	case msg := <-consumers:
		if msg.Topic != client.tradeTopic {
			t.Errorf("Expected topic %v, got %v", client.tradeTopic, msg.Topic)
		}
		var received Trade
		if err := json.Unmarshal(msg.Value, &received); err != nil {
			t.Fatalf("Failed to unmarshal consumed message: %v", err)
		}
		if received.OrderId != trade.OrderId {
			t.Errorf("Expected OrderId %v, got %v", trade.OrderId, received.OrderId)
		}
		if received.Quantity != trade.Quantity {
			t.Errorf("Expected Quantity %v, got %v", trade.Quantity, received.Quantity)
		}
		if received.Price != trade.Price {
			t.Errorf("Expected Price %v, got %v", trade.Price, received.Price)
		}
		if received.Action != trade.Action {
			t.Errorf("Expected Action %v, got %v", trade.Action, received.Action)
		}
		if received.Status != trade.Status {
			t.Errorf("Expected Status %v, got %v", trade.Status, received.Status)
		}
	case errMsg := <-errors:
		t.Fatalf("Received unexpected error: %v", errMsg.Err)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

func TestSendPricePoint(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()

	mockConsumer, ok := client.MockConsumer.(*mocks.Consumer)
	if !ok {
		t.Fatal("Expected MockConsumer to be *mocks.Consumer")
	}

	pc := mockConsumer.ExpectConsumePartition(client.pricePointTopic, 0, sarama.OffsetOldest)
	defer pc.Close()

	consumers, errors := client.Assign(*consumer, client.pricePointTopic)

	pp := PricePoint{
		Price: 50.0,
	}
	client.MockChan.SendPricePoint(pp)
	close(client.MockChan)

	select {
	case msg := <-consumers:
		if msg.Topic != client.pricePointTopic {
			t.Errorf("Expected topic %v, got %v", client.pricePointTopic, msg.Topic)
		}
		var received PricePoint
		if err := json.Unmarshal(msg.Value, &received); err != nil {
			t.Fatalf("Failed to unmarshal consumed message: %v", err)
		}
		if received.Price != pp.Price {
			t.Errorf("Expected Price %v, got %v", pp.Price, received.Price)
		}
	case errMsg := <-errors:
		t.Fatalf("Received unexpected error: %v", errMsg.Err)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

func TestAssign(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()

	mockConsumer, ok := client.MockConsumer.(*mocks.Consumer)
	if !ok {
		t.Fatal("Expected MockConsumer to be *mocks.Consumer")
	}

	pc := mockConsumer.ExpectConsumePartition(client.quoteTopic, 0, sarama.OffsetOldest)
	defer pc.Close()

	consumers, errors := client.Assign(*consumer, client.quoteTopic)

	order := Order{
		OrderID:   uuid.New().String(),
		OrderType: "limit",
		Price:     50.5,
		Quantity:  100,
		Timestamp: float64(time.Now().Unix()),
	}
	client.MockChan.SendOrder(order)

	trade := Trade{
		OrderId:  uuid.New().String(),
		Quantity: 200,
		Price:    60.0,
		Action:   "sell",
		Status:   "",
	}
	client.MockChan.SendTrade(trade)

	close(client.MockChan)

	select {
	case msg := <-consumers:
		if msg.Topic != client.quoteTopic {
			t.Errorf("Expected topic %v, got %v", client.quoteTopic, msg.Topic)
		}
		var received Order
		if err := json.Unmarshal(msg.Value, &received); err != nil {
			t.Fatalf("Failed to unmarshal consumed message: %v", err)
		}
		if received.OrderID != order.OrderID {
			t.Errorf("Expected OrderID %v, got %v", order.OrderID, received.OrderID)
		}
		if received.Price != order.Price {
			t.Errorf("Expected Price %v, got %v", order.Price, received.Price)
		}
	case errMsg := <-errors:
		t.Fatalf("Received unexpected error: %v", errMsg.Err)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}

	if msg, ok := <-consumers; ok {
		t.Errorf("Received unexpected message: %v", msg)
	} else {
	}
}

func TestProduceConsumeEndToEnd(t *testing.T) {
	client := NewMockKafkaClient(t)
	consumer := client.GetConsumer()
	producer := client.GetProducer()

	mockConsumer, ok := client.MockConsumer.(*mocks.Consumer)
	if !ok {
		t.Fatal("Expected MockConsumer to be *mocks.Consumer")
	}

	pc := mockConsumer.ExpectConsumePartition(client.quoteTopic, 0, sarama.OffsetOldest)
	defer pc.Close()

	consumers, errors := client.Assign(*consumer, client.quoteTopic)

	// Création d'un message Order à produire
	order := Order{
		OrderID:   uuid.New().String(),
		OrderType: "limit",
		Price:     50.5,
		Quantity:  100,
		Timestamp: float64(time.Now().Unix()),
	}
	orderBytes, _ := json.Marshal(order)
	producerMsg := &sarama.ProducerMessage{
		Topic: client.quoteTopic,
		Value: sarama.ByteEncoder(orderBytes),
	}

	mockProducer, ok := client.MockProducer.(*mocks.SyncProducer)
	if !ok {
		t.Fatal("Expected MockProducer to be *mocks.SyncProducer")
	}
	mockProducer.ExpectSendMessageAndSucceed()

	partition, _, err := (*producer).SendMessage(producerMsg)
	if err != nil {
		t.Fatalf("Failed to produce message: %v", err)
	}
	if partition != 0 {
		t.Errorf("Expected partition 0, got %v", partition)
	}
	client.MockChan.SendOrder(order)
	close(client.MockChan)

	select {
	case msg := <-consumers:
		if msg.Topic != client.quoteTopic {
			t.Errorf("Expected topic %v, got %v", client.quoteTopic, msg.Topic)
		}
		var received Order
		if err := json.Unmarshal(msg.Value, &received); err != nil {
			t.Fatalf("Failed to unmarshal consumed message: %v", err)
		}
		if received.OrderID != order.OrderID {
			t.Errorf("Expected OrderID %v, got %v", order.OrderID, received.OrderID)
		}
		if received.Price != order.Price {
			t.Errorf("Expected Price %v, got %v", order.Price, received.Price)
		}
	case errMsg := <-errors:
		t.Fatalf("Received unexpected error: %v", errMsg.Err)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}
