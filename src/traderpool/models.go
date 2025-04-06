package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  int64   `json:"quantity"`
	Timestamp float64 `json:"timestamp"`
}

type Trade struct {
	OrderId  string  `json:"order_id"`
	Quantity int64   `json:"quantity"`
	Price    float64 `json:"price"`
	Action   string  `json:"action"`
	Status   string  `json:"status"`
}

type PricePoint struct {
	Price float64 `json:"price"`
}

type Trader struct {
	TraderId           string
	Quantity           int64
	AggressivenessBuy  float64
	AggressivenessSell float64
	TargetBuy          float64
	TargetSell         float64
	Sign               int64

	MidPrice *MidPrice

	KafkaClient *KafkaClient

	QuoteTopic      string
	TradeTopic      string
	PricePointTopic string
}

type MidPrice struct {
	Value float64 `json:"value"`
	sync.Mutex
}

func createMidPrice(midPrice float64) *MidPrice {
	return &MidPrice{
		Value: midPrice,
	}
}

func NewTrader(traderId int, midPrice *MidPrice, kafkaClient *KafkaClient) *Trader {
	t := new(Trader)
	t.TraderId = strconv.Itoa(traderId)
	t.KafkaClient = kafkaClient

	randomNumber := rand.Float64()
	t.AggressivenessBuy = 3 * randomNumber
	t.AggressivenessSell = 3 * randomNumber

	randomQuantity := rand.Int63n(50) + 1
	if traderId%2 == 0 {
		t.Sign = 1
		t.Quantity = randomQuantity
	} else {
		t.Sign = -1
		t.Quantity = -randomQuantity
	}

	t.MidPrice = midPrice

	fmt.Printf("Trader %s, Agg Buy: %.3f, Agg Sell: %.3f\n",
		t.TraderId, t.AggressivenessBuy, t.AggressivenessSell)

	t.QuoteTopic = "orders.topic"
	t.TradeTopic = "trades.topic"
	t.PricePointTopic = "order.last_price.topic"
	return t
}
