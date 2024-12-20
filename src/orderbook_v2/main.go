package main

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/IBM/sarama"
)

var topic string = ""

type Order struct {
}

type KafkaClient struct {
	config   sarama.Config
	consumer sarama.ConsumerGroup
}

func NewKafkaClient() *KafkaClient {
	kc := new(KafkaClient)
	kc.config = *sarama.NewConfig()
	return kc
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
