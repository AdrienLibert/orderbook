package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"math/rand"
	"time"
)

func GenerateAndPushOrder(trader Trader, orderChannel chan<- Order) {
	orderID := fmt.Sprintf("%s-%d", trader.TradeId, time.Now().UnixNano())
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	quantity := int64(10 + rng.Intn(30))
	if rng.Float64() < 0.5 {
		quantity = -quantity
	}

	adjustedPrice := trader.Price
	if quantity > 0 {
		if rng.Float64() < 0.3 {
			adjustedPrice = trader.Price * (1 + rng.Float64()*0.01)
		} else {
			adjustedPrice = trader.Price * (1 - rng.Float64()*0.01)
		}
	} else {
		if rng.Float64() < 0.3 {
			adjustedPrice = trader.Price * (1 - rng.Float64()*0.01)
		} else {
			adjustedPrice = trader.Price * (1 + rng.Float64()*0.01)
		}
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
	fmt.Printf("INFO: Trader %s generated order from price %.2f: %+v\n", trader.TradeId, trader.Price, order)
}

func StartTrader(traderID string, priceChannel <-chan Trader, orderChannel chan<- Order) {
	go func() {
		for trader := range priceChannel {
			if trader.TradeId == traderID {
				fmt.Println("test")
				GenerateAndPushOrder(trader, orderChannel)
			}
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
