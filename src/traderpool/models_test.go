package main

import (
	"sync"
	"testing"
)

func TestNewTrader(t *testing.T) {
	midPrice := createMidPrice(100.0)
	kafkaClient := &KafkaClient{}

	trader1 := NewTrader(2, midPrice, kafkaClient)
	if trader1.TraderId != "2" {
		t.Errorf("Expected TraderId '2', got '%s'", trader1.TraderId)
	}
	if trader1.Sign != 1 {
		t.Errorf("Expected Sign 1 for even traderId, got %d", trader1.Sign)
	}
	if trader1.Quantity <= 0 || trader1.Quantity > 50 {
		t.Errorf("Expected Quantity between 1 and 50, got %d", trader1.Quantity)
	}
	if trader1.AggressivenessBuy < 0 || trader1.AggressivenessBuy > 3 {
		t.Errorf("Expected AggressivenessBuy between 0 and 3, got %.3f", trader1.AggressivenessBuy)
	}
	if trader1.AggressivenessSell < 0 || trader1.AggressivenessSell > 3 {
		t.Errorf("Expected AggressivenessSell between 0 and 3, got %.3f", trader1.AggressivenessSell)
	}
	if trader1.MidPrice != midPrice {
		t.Errorf("Expected MidPrice to match, got different pointer")
	}
	if trader1.QuoteTopic != "orders.topic" {
		t.Errorf("Expected QuoteTopic 'orders.topic', got '%s'", trader1.QuoteTopic)
	}
	if trader1.TradeTopic != "trades.topic" {
		t.Errorf("Expected TradeTopic 'trades.topic', got '%s'", trader1.TradeTopic)
	}
	if trader1.PricePointTopic != "order.last_price.topic" {
		t.Errorf("Expected PricePointTopic 'order.last_price.topic', got '%s'", trader1.PricePointTopic)
	}

	trader2 := NewTrader(3, midPrice, kafkaClient)
	if trader2.TraderId != "3" {
		t.Errorf("Expected TraderId '3', got '%s'", trader2.TraderId)
	}
	if trader2.Sign != -1 {
		t.Errorf("Expected Sign -1 for odd traderId, got %d", trader2.Sign)
	}
	if trader2.Quantity >= 0 || trader2.Quantity < -50 {
		t.Errorf("Expected Quantity between -50 and -1, got %d", trader2.Quantity)
	}
}

func TestCreateMidPrice(t *testing.T) {
	midPrice := createMidPrice(150.5)
	if midPrice.Value != 150.5 {
		t.Errorf("Expected MidPrice.Value 150.5, got %.2f", midPrice.Value)
	}
	midPrice.Lock()
	midPrice.Value = 200.0
	midPrice.Unlock()
	if midPrice.Value != 200.0 {
		t.Errorf("Expected MidPrice.Value 200.0 after update, got %.2f", midPrice.Value)
	}
}

func TestMidPriceConcurrency(t *testing.T) {
	midPrice := createMidPrice(100.0)
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(delta float64) {
			defer wg.Done()
			midPrice.Lock()
			midPrice.Value += delta
			midPrice.Unlock()
		}(1.0)
	}

	wg.Wait()
	expected := 100.0 + float64(numGoroutines)
	if midPrice.Value != expected {
		t.Errorf("Expected MidPrice.Value %.2f after concurrent updates, got %.2f", expected, midPrice.Value)
	}
}
