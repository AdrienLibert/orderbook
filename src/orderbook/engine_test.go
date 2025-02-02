package main

import (
	"testing"
	"time"
)

func TestMachingEngineProcessScenarios(t *testing.T) {
	orderbook := NewOrderBook()
	machingEngine := NewMatchingEngine(nil, orderbook)

	now := time.Now().UTC().Unix()

	// 1. Add a buy order
	buyPrice := 10.0
	buyOrder := Order{
		OrderID:   "buy-uuid-uuid-uuid",
		OrderType: "buy",
		Price:     buyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&buyOrder, nil, nil)
	if orderbook.BestBid.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestBid.Len(), 1)
	}
	if orderbook.BestBid.Peak() != buyPrice {
		t.Errorf("got %f, wanted %f", orderbook.BestBid.Peak(), buyPrice)
	}
	if len(orderbook.PriceToBuyOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToBuyOrders), 1)
	}

	if orderbook.BestAsk.Len() != 0 {
		t.Errorf("got %d, wanted %d", orderbook.BestAsk.Len(), 0)
	}
	if orderbook.BestAsk.Peak() != nil {
		t.Errorf("got %f, wanted %s", orderbook.BestAsk.Peak(), "nil")
	}
	if len(orderbook.PriceToSellOrders) != 0 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToSellOrders), 0)
	}

	// 2. Add a sell order with higher price
	sellPrice := 11.0
	sellOrder := Order{
		OrderID:   "sell-uuid-uuid-uuid",
		OrderType: "sell",
		Price:     sellPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&sellOrder, nil, nil)
	if orderbook.BestAsk.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestAsk.Len(), 1)
	}
	if orderbook.BestAsk.Peak() != sellPrice {
		t.Errorf("got %f, wanted %f", orderbook.BestAsk.Peak(), 1.0)
	}
	if len(orderbook.PriceToSellOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToSellOrders), 1)
	}

	if orderbook.BestBid.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestBid.Len(), 1)
	}
	if orderbook.BestBid.Peak() != buyPrice {
		t.Errorf("got %f, wanted %f", orderbook.BestBid.Peak(), buyPrice)
	}
	if len(orderbook.PriceToBuyOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToBuyOrders), 1)
	}

	// 3. Add a matching buy Order which empties the sell side
	newBuyPrice := 11.0
	newBuyOrder := Order{
		OrderID:   "new-buy-uuid-uuid",
		OrderType: "buy",
		Price:     newBuyPrice,
		Quantity:  20.0,
		Timestamp: now,
	}
	machingEngine.Process(&newBuyOrder, nil, nil)
	if orderbook.BestBid.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestBid.Len(), 1)
	}
	if orderbook.BestBid.Peak() != buyPrice {
		t.Errorf("got %f, wanted %f", orderbook.BestBid.Peak(), buyPrice)
	}
	if len(orderbook.PriceToBuyOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToBuyOrders), 1)
	}

	if orderbook.BestAsk.Len() != 0 {
		t.Errorf("got %d, wanted %d", orderbook.BestAsk.Len(), 0)
	}
	if orderbook.BestAsk.Peak() != nil {
		t.Errorf("got %f, wanted %s", orderbook.BestAsk.Peak(), "nil")
	}
	if len(orderbook.PriceToSellOrders) != 0 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToSellOrders), 0)
	}
}
