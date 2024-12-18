package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Order struct {
}

type Orderbook struct {
	bid map[decimal.Decimal][]Order
	ask map[decimal.Decimal][]Order
}

func SimpleOrderBook() *Orderbook {
	o := new(Orderbook)
	o.bid = make(map[decimal.Decimal][]Order)
	o.ask = make(map[decimal.Decimal][]Order)

	return o
}

func (o Orderbook) Start() {
	fmt.Println("Starting orderbook")
	for {
		fmt.Println("Running orderbook")
	}
}

func main() {
	fmt.Println("Starting orderbook")
	o := SimpleOrderBook()
	o.Start()
}
