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
	buyPrice := 10.0
	buyOrder := Order{
		OrderID:   "buy-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     buyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&buyOrder, nil, nil)
	// One element in buy book
	assert.Equal(t, orderbook.BestBid.Len(), 1)
	assert.Equal(t, orderbook.BestBid.Peak(), buyPrice)
	assert.Equal(t, len(orderbook.PriceToBuyOrders), 1)
	assert.Equal(t, orderbook.PriceToBuyOrders[buyPrice], []*Order{&buyOrder})
	// No element in sell book
	assert.Equal(t, orderbook.BestAsk.Len(), 0)
	assert.Equal(t, orderbook.BestAsk.Peak(), nil)
	assert.Equal(t, len(orderbook.PriceToSellOrders), 0)

	// 2. Add a sell order with higher price
	sellPrice := 11.0
	sellOrder := Order{
		OrderID:   "sell-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     sellPrice,
		Quantity:  -20.0,
		Timestamp: now,
	}
	machingEngine.Process(&sellOrder, nil, nil)
	// One element in buy book
	assert.Equal(t, orderbook.BestBid.Len(), 1)
	assert.Equal(t, orderbook.BestBid.Peak(), buyPrice)
	assert.Equal(t, len(orderbook.PriceToBuyOrders), 1)
	assert.Equal(t, orderbook.PriceToBuyOrders[buyPrice], []*Order{&buyOrder})

	// One element in sell book
	assert.Equal(t, orderbook.BestAsk.Len(), 1)
	assert.Equal(t, orderbook.BestAsk.Peak(), sellPrice)
	assert.Equal(t, len(orderbook.PriceToSellOrders), 1)
	assert.Equal(t, orderbook.PriceToSellOrders[sellPrice], []*Order{&sellOrder})

	// 3. Add a matching buy Order which empties the sell side
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
	assert.Equal(t, orderbook.BestBid.Len(), 1)
	assert.Equal(t, orderbook.BestBid.Peak(), buyPrice)
	assert.Equal(t, len(orderbook.PriceToBuyOrders), 1)
	assert.Equal(t, orderbook.PriceToBuyOrders[buyPrice], []*Order{&buyOrder})

	// No more element in sell book
	assert.Equal(t, orderbook.BestAsk.Len(), 0)
	assert.Equal(t, orderbook.BestAsk.Peak(), nil)
	assert.Equal(t, len(orderbook.PriceToSellOrders), 0)
}
