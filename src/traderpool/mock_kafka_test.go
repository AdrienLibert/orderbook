package main

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"testing"
)

// MockConsumerChan for testing
type MockConsumerChan chan *sarama.ConsumerMessage

func (m MockConsumerChan) SendOrder(orderId, status string) {
	trade := Trade{
		OrderId:  orderId,
		Status:   status,
		Quantity: 0,
		Price:    0.0,
		Action:   "",
	}
	tradeBytes, _ := json.Marshal(trade)
	m <- &sarama.ConsumerMessage{
		Value: tradeBytes,
	}
}

// MockKafkaClient for testing
type MockKafkaClient struct {
	*KafkaClient
	MockProducer sarama.SyncProducer
}

func NewMockKafkaClient(t *testing.T) *MockKafkaClient {
	kc := NewKafkaClient()
	mockProducer := mocks.NewSyncProducer(t, kc.producerConfig)
	return &MockKafkaClient{
		KafkaClient:  kc,
		MockProducer: mockProducer,
	}
}

func (mkc *MockKafkaClient) GetProducer() *sarama.SyncProducer {
	return &mkc.MockProducer
}
