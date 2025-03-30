package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp float64 `json:"timestamp"`
}

type Trade struct {
	OrderId  string  `json:"order_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Action   string  `json:"action"`
	Status   string  `json:"status"`
}

type PricePoint struct {
	Price float64 `json:"price"`
}

type Trader struct {
	traderId           string
	quantity           float64
	aggressivenessBuy  float64
	aggressivenessSell float64
	targetBuy          float64
	targetSell         float64

	midPrice *MidPrice

	kafkaClient *KafkaClient

	_quoteTopic      string
	_tradeTopic      string
	_pricePointTopic string
}

type MidPrice struct {
	Value  float64 `json:"value"`
	Spread float64 `json:"spread"`
	sync.Mutex
}

func createMidPrice(midPrice float64, spread float64) *MidPrice {
	return &MidPrice{
		Value:  midPrice,
		Spread: spread,
	}
}

func NewTrader(traderId string, quantity float64, midPrice *MidPrice, kafkaClient *KafkaClient) *Trader {
	t := new(Trader)
	t.traderId = traderId
	t.quantity = quantity
	t.kafkaClient = kafkaClient

	randomNumber := rand.Float64()
	t.aggressivenessBuy = -0.15 + 0.3*randomNumber
	t.aggressivenessSell = -0.15 + 0.3*randomNumber

	t.midPrice = midPrice

	fmt.Printf("Trader %s -> Limit Price: %.2f, Agg Buy: %.3f, Agg Sell: %.3f\n",
		t.traderId, t.aggressivenessBuy, t.aggressivenessSell)

	t._quoteTopic = "orders.topic"
	t._tradeTopic = "trades.topic"
	t._pricePointTopic = "order.last_price.topic"
	return t
}
