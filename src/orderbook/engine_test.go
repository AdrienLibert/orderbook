package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMachingEngineProcessIterative(t *testing.T) {
	orderbook := NewOrderBook()
	machingEngine := NewMatchingEngine(nil, orderbook)

	now := time.Now().UTC().Unix()

	// 1. Add a buy order
	firstBuyPrice := 10.0
	firstBuyOrder := Order{
		OrderID:   "buy-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     firstBuyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&firstBuyOrder, nil, nil)
	// One element in buy book
	assert.Equal(t, 1, orderbook.BestBid.Len())
	assert.Equal(t, firstBuyPrice, orderbook.BestBid.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToBuyOrders))
	assert.Equal(t, &[]*Order{&firstBuyOrder}, orderbook.PriceToBuyOrders[firstBuyPrice])
	// No element in sell book
	assert.Equal(t, 0, orderbook.BestAsk.Len())
	assert.Equal(t, nil, orderbook.BestAsk.Peak())
	assert.Equal(t, 0, len(orderbook.PriceToSellOrders))

	// 2. Add a sell order with higher price
	sellPrice := 11.0
	sellOrder := Order{
		OrderID:   "sell-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     sellPrice,
		Quantity:  -25.0,
		Timestamp: now,
	}
	machingEngine.Process(&sellOrder, nil, nil)
	// One element in buy book
	assert.Equal(t, 1, orderbook.BestBid.Len())
	assert.Equal(t, firstBuyPrice, orderbook.BestBid.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToBuyOrders))
	assert.Equal(t, &[]*Order{&firstBuyOrder}, orderbook.PriceToBuyOrders[firstBuyPrice])

	// One element in sell book
	assert.Equal(t, 1, orderbook.BestAsk.Len())
	assert.Equal(t, sellPrice, orderbook.BestAsk.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToSellOrders))
	assert.Equal(t, &[]*Order{&sellOrder}, orderbook.PriceToSellOrders[sellPrice])

	// 3. Add a matching buy Order which reduces the sell side
	newBuyPrice := 11.0
	newBuyOrder := Order{
		OrderID:   "new-buy-uuid-uuid",
		OrderType: "limit",
		Price:     newBuyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&newBuyOrder, nil, nil)
	// One element in buy book
	assert.Equal(t, 1, orderbook.BestBid.Len())
	assert.Equal(t, firstBuyPrice, orderbook.BestBid.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToBuyOrders))
	assert.Equal(t, &[]*Order{&firstBuyOrder}, orderbook.PriceToBuyOrders[firstBuyPrice])

	// One element in sell book (but partially filled)
	assert.Equal(t, 1, orderbook.BestAsk.Len())
	assert.Equal(t, sellPrice, orderbook.BestAsk.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToSellOrders))
	assert.Equal(t, &[]*Order{&sellOrder}, orderbook.PriceToSellOrders[sellPrice])

	// 4. Empty sell side with last buy Order
	lastBuyPrice := 11.0
	lastBuyOrder := Order{
		OrderID:   "new-buy-uuid-uuid",
		OrderType: "limit",
		Price:     lastBuyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&lastBuyOrder, nil, nil)
	// Two element in buy book: the first and the last (partial)
	assert.Equal(t, 2, orderbook.BestBid.Len())
	assert.Equal(t, firstBuyPrice, orderbook.BestBid.Peak())
	assert.Equal(t, 2, len(orderbook.PriceToBuyOrders))
	assert.Equal(t, &[]*Order{&firstBuyOrder}, orderbook.PriceToBuyOrders[firstBuyPrice])

	// No element in sell book
	assert.Equal(t, 0, orderbook.BestAsk.Len())
	assert.Equal(t, nil, orderbook.BestAsk.Peak())
	assert.Equal(t, 0, len(orderbook.PriceToSellOrders))
}
