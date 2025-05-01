package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"math/rand"
	"time"
)

func GenerateAndPushOrder(traderID string, price float64, orderChannel chan<- Order) {
	orderID := fmt.Sprintf("%s-%d", traderID, time.Now().UnixNano())

	quantity := int64(10 + rand.Intn(30))
	if rand.Float64() < 0.5 {
		quantity = -quantity
	}

	adjustedPrice := price
	switch {
	case rand.Float64() < 0.25:
		adjustedPrice = price * 1.05
	case rand.Float64() < 0.5:
		adjustedPrice = price * 1.01
	case rand.Float64() < 0.75:
		adjustedPrice = price * 0.99
	default:
		adjustedPrice = price * 0.95
	}

	timestamp := int64(time.Now().UnixNano()) / 1e9

	order := Order{
		OrderID:   orderID,
		OrderType: "limit",
		Price:     adjustedPrice,
		Quantity:  quantity,
		Timestamp: timestamp,
	}

	orderChannel <- order
	fmt.Printf("INFO: Trader %s generated order from price %.2f: %+v\n", traderID, price, order)
}

func StartTrader(traderID string, priceChannel <-chan float64, orderChannel chan<- Order) {
	go func() {
		for price := range priceChannel {
			GenerateAndPushOrder(traderID, price, orderChannel)
		}
		fmt.Printf("INFO: Trader %s stopped, priceListChannel closed\n", traderID)
	}()
}

func RequestMidPrice(priceConsumer <-chan *sarama.ConsumerMessage) float64 {
	var midPrice float64
	select {
	case priceMsg := <-priceConsumer:
		pricePoint, err := convertMessageToPricePoint(priceMsg.Value)
		if err != nil {
			fmt.Printf("ERROR: Failed to parse PricePoint: %v\n", err)
			midPrice = 50.0
		} else {
			midPrice = pricePoint.Price
			fmt.Printf("INFO: MidPrice set to: %.2f\n", midPrice)
		}
	default:
		fmt.Println("INFO: No price message available, using default MidPrice")
		midPrice = 50.0
	}
	return midPrice
}
