package main

type Heap interface {
	Push(x interface{})
	Pop() interface{}
	Peak() interface{}
	Len() int
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
	BestBid       *MaxHeap
	BestAsk       *MinHeap
	PriceToVolume map[float64]float64
	// indexes + containers
	PriceToBuyOrders  map[float64][]*Order
	PriceToSellOrders map[float64][]*Order
}

func NewOrderBook() *Orderbook {
	o := new(Orderbook)
	o.BestBid = &MaxHeap{}
	o.BestAsk = &MinHeap{}
	o.PriceToBuyOrders = make(map[float64][]*Order)
	o.PriceToSellOrders = make(map[float64][]*Order)
	return o
}

func (o *Orderbook) AddOrder(order *Order, orderType string) {
	// Add order in orderbook
	// Rules:
	// - Hashmap of orders are indexes used to assess price in heaps exist
	// - Orders are added at the end of the list of orders
	price := order.Price

	if orderType == "BUY" {
		val, ok := o.PriceToBuyOrders[price]
		if ok {
			o.PriceToBuyOrders[price] = append(val, order)
		} else {
			o.BestBid.Push(price)
			o.PriceToBuyOrders[price] = []*Order{order}
		}
	}
	if orderType == "SELL" {
		val, ok := o.PriceToSellOrders[price]
		if ok {
			o.PriceToSellOrders[price] = append(val, order)
		} else {
			o.BestAsk.Push(price)
			o.PriceToSellOrders[price] = []*Order{order}
		}
	}
}
