package main

import (
	"fmt"
	"os"
)

// utils
func cut(i int, xs *[]*Order) *Order {
	y := (*xs)[i]
	*xs = append((*xs)[:i], (*xs)[i+1:]...)
	return y
}

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// TODO: move to models file

func handleError(err error) {
	fmt.Println("ERROR: invalid message consummed:", err)
}

func publishTrade(inOrder *Order, tradeQuantity float64, price float64, action string) Trade {
	var status string

	if inOrder.Quantity == 0 {
		status = "closed"
	} else {
		status = "partial"
	}

	trade := Trade{
		OrderId:  inOrder.OrderID,
		Quantity: tradeQuantity,
		Price:    price,
		Action:   action,
		Status:   status,
	}

	return trade
}

func publishPricePoint(price float64) PricePoint {
	return PricePoint{
		Price: price,
	}
}

func main() {
	fmt.Println("INFO: tarting orderbook")
	me := NewMatchingEngine(
		NewKafkaClient(),
		NewOrderBook(),
	)
	me.Start()
}
