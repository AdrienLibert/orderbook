package main

import (
	"container/heap"
	"testing"
	"time"
)

func TestMinHeapPop(t *testing.T) {
	h := &MinHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{0.0, 1.0, 2.0, 5.0, 6.0, 7.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := heap.Pop(h)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
	}
}

func TestMinHeapPeak(t *testing.T) {
	h := &MinHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{0.0, 1.0, 2.0, 5.0, 6.0, 7.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := h.Peak().(float64)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
		heap.Pop(h)
	}
}

func TestMaxHeapPop(t *testing.T) {
	h := &MaxHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{7.0, 6.0, 5.0, 2.0, 1.0, 0.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := heap.Pop(h)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
	}
}

func TestMaxHeapPeak(t *testing.T) {
	h := &MaxHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{7.0, 6.0, 5.0, 2.0, 1.0, 0.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := h.Peak().(float64)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
		heap.Pop(h)
	}
}

func TestOrderBookAddOrder(t *testing.T) {
	orderbook := NewOrderBook()
	now := time.Now().UTC().Unix()
	price := 10.0
	buyOrder := Order{
		OrderID:   "uuid-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     price,
		Quantity:  20.0,
		Timestamp: now,
	}

	orderbook.AddOrder(&buyOrder, "BUY")
	if orderbook.BestBid.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestBid.Len(), 1)
	}
	if orderbook.BestBid.Peak() != price {
		t.Errorf("got %f, wanted %f", orderbook.BestBid.Peak(), price)
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

	sellOrder := Order{
		OrderID:   "uuid-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     price,
		Quantity:  -20.0,
		Timestamp: now,
	}
	orderbook.AddOrder(&sellOrder, "SELL")
	if orderbook.BestAsk.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestAsk.Len(), 1)
	}
	if orderbook.BestAsk.Peak() != price {
		t.Errorf("got %f, wanted %f", orderbook.BestAsk.Peak(), price)
	}
	if len(orderbook.PriceToSellOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToSellOrders), 1)
	}

	if orderbook.BestBid.Len() != 1 {
		t.Errorf("got %d, wanted %d", orderbook.BestBid.Len(), 1)
	}
	if orderbook.BestBid.Peak() != price {
		t.Errorf("got %f, wanted %f", orderbook.BestBid.Peak(), price)
	}
	if len(orderbook.PriceToBuyOrders) != 1 {
		t.Errorf("got %d, wanted %d", len(orderbook.PriceToBuyOrders), 1)
	}
}
