package main

import (
	"encoding/json"
	"fmt"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp int64   `json:"timestamp"`
}

type Trade struct {
	TradeId   string  `json:"trade_id"`
	OrderId   string  `json:"order_id"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Action    string  `json:"action"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
}

type PricePoint struct {
	Price float64 `json:"price"`
}

func (trade Trade) toJSON() []byte {
	message, err := json.Marshal(trade)
	if err != nil {
		fmt.Println("ERROR: invalid trade being converted to message:", err)
	}
	return message
}

func (pricePoint PricePoint) toJSON() []byte {
	message, err := json.Marshal(pricePoint)
	if err != nil {
		fmt.Println("ERROR: invalid price point being converted to message:", err)
	}
	return message
}

func messageToOrder(messageValue []byte) (Order, error) {
	var order = &Order{}
	if err := json.Unmarshal(messageValue, order); err != nil {
		return Order{}, err
	}
	return *order, nil
}

func createTrade(tradeId string, inOrder *Order, tradeQuantity float64, price float64, action string, ts int64) Trade {
	var status string

	if inOrder.Quantity == 0 {
		status = "closed"
	} else {
		status = "partial"
	}

	trade := Trade{
		TradeId:   tradeId,
		OrderId:   inOrder.OrderID,
		Quantity:  tradeQuantity,
		Price:     price,
		Action:    action,
		Status:    status,
		Timestamp: ts,
	}

	return trade
}

func createPricePoint(price float64) PricePoint {
	return PricePoint{
		Price: price,
	}
}
