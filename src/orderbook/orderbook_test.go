package main

import (
	"container/heap"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 1, orderbook.BestBid.Len())
	assert.Equal(t, price, orderbook.BestBid.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToBuyOrders))

	sellOrder := Order{
		OrderID:   "uuid-uuid-uuid-uuid",
		OrderType: "limit",
		Price:     price,
		Quantity:  -20.0,
		Timestamp: now,
	}
	orderbook.AddOrder(&sellOrder, "SELL")
	assert.Equal(t, 1, orderbook.BestAsk.Len())
	assert.Equal(t, price, orderbook.BestAsk.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToSellOrders))

	assert.Equal(t, 1, orderbook.BestBid.Len())
	assert.Equal(t, price, orderbook.BestBid.Peak())
	assert.Equal(t, 1, len(orderbook.PriceToBuyOrders))
}
