package main

import (
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestTargetPrices(t *testing.T) {
	trader := &Trader{
		MidPrice:           &MidPrice{Value: 100.0},
		AggressivenessBuy:  2.0,
		AggressivenessSell: 1.5,
	}
	trader.TargetPrices()

	expectedBuy := 100.0 + (2.0 * 0.5)
	expectedSell := 100.0 - (1.5 * 0.5)
	if trader.TargetBuy != expectedBuy {
		t.Errorf("Expected TargetBuy %.2f, got %.2f", expectedBuy, trader.TargetBuy)
	}
	if trader.TargetSell != expectedSell {
		t.Errorf("Expected TargetSell %.2f, got %.2f", expectedSell, trader.TargetSell)
	}
}

func TestTrade(t *testing.T) {
	orderChan := make(chan Order, 10)
	trader := &Trader{
		MidPrice:           &MidPrice{Value: 100.0},
		AggressivenessBuy:  2.0,
		AggressivenessSell: 1.5,
		Sign:               1,
		TraderId:           "1",
	}

	order := trader.Trade(orderChan)
	if order.Quantity <= 0 || order.Quantity > 50 {
		t.Errorf("Expected Quantity between 1 and 50, got %d", order.Quantity)
	}
	if order.Price != trader.TargetBuy {
		t.Errorf("Expected Price %.2f (TargetBuy) for positive Quantity, got %.2f", trader.TargetBuy, order.Price)
	}

	trader.Sign = -1
	order = trader.Trade(orderChan)
	if order.Quantity >= 0 || order.Quantity < -50 {
		t.Errorf("Expected Quantity between -50 and -1, got %d", order.Quantity)
	}
	if order.Price != trader.TargetSell {
		t.Errorf("Expected Price %.2f (TargetSell) for negative Quantity, got %.2f", trader.TargetSell, order.Price)
	}

	select {
	case received := <-orderChan:
		if received.Quantity == 0 {
			t.Errorf("Received empty order in channel")
		}
	default:
		t.Errorf("Order not sent to channel")
	}
}

func TestTradeChannelFull(t *testing.T) {
	orderChan := make(chan Order, 1)
	trader := &Trader{
		MidPrice:           &MidPrice{Value: 100.0},
		AggressivenessBuy:  2.0,
		AggressivenessSell: 1.5,
		Sign:               1,
	}

	orderChan <- Order{Quantity: 1, Price: 100.0}
	order := trader.Trade(orderChan)
	if order.Quantity == 0 {
		t.Errorf("Expected non-zero Quantity, got %d", order.Quantity)
	}
}

func TestStart(t *testing.T) {
	kc := NewKafkaClient()
	trader := &Trader{
		MidPrice:           &MidPrice{Value: 100.0},
		AggressivenessBuy:  2.0,
		AggressivenessSell: 1.5,
		Sign:               1,
		TraderId:           "1",
		KafkaClient:        kc,
		QuoteTopic:         "quotes",
		TradeTopic:         "trades",
		PricePointTopic:    "prices",
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		signals <- os.Interrupt
	}()

	go trader.Start()
}
