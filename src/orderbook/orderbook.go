package main

type Heap interface {
	Push(x interface{})
	Pop() interface{}
	Peak() interface{}
}

type MinHeap []float64

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] }

func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) { *h = append(*h, x.(float64)) }

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *MinHeap) Peak() interface{} {
	n := len(*h)
	if n == 0 {
		return nil
	}
	return (*h)[0]
}

type MaxHeap []float64

func (h MaxHeap) Len() int { return len(h) }

func (h MaxHeap) Less(i, j int) bool { return h[i] > h[j] }

func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) { *h = append(*h, x.(float64)) }

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *MaxHeap) Peak() interface{} {
	n := len(*h)
	if n == 0 {
		return nil
	}
	return (*h)[0]
}

type Orderbook struct {
	BestBid           *MaxHeap
	BestAsk           *MinHeap
	PriceToVolume     map[float64]float64
	PriceToSellOrders map[float64][]*Order
	PriceToBuyOrders  map[float64][]*Order
}

func NewOrderBook() *Orderbook {
	o := new(Orderbook)
	o.BestBid = &MaxHeap{}
	o.BestAsk = &MinHeap{}
	return o
}
