package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type MatchingEngine struct {
	kafkaClient     *KafkaClient
	orderBook       *Orderbook
	quoteTopic      string
	tradeTopic      string
	pricePointTopic string
}

func NewMatchingEngine(kafkaClient *KafkaClient, orderBook *Orderbook) *MatchingEngine {
	me := new(MatchingEngine)
	me.kafkaClient = kafkaClient
	me.orderBook = orderBook
	me.quoteTopic = "orders.topic"
	me.tradeTopic = "trades.topic"
	me.pricePointTopic = "order.last_price.topic"
	return me
}

func (me *MatchingEngine) Start() {
	orderMessages, errors := me.kafkaClient.ConsumerMessagesChan(me.quoteTopic)

	tradeProducer := me.kafkaClient.GetProducer()
	pricePointProducer := me.kafkaClient.GetProducer()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	tradeChannel := make(chan Trade)
	go func(tradeMessages <-chan Trade) {
		for {
			select {
			case trade := <-tradeMessages:
				producerMessage := sarama.ProducerMessage{Topic: me.tradeTopic, Value: sarama.StringEncoder(trade.toJSON())}
				par, off, err := (*tradeProducer).SendMessage(&producerMessage)
				if err != nil {
					fmt.Printf("ERROR: producing trade in partition %d, offset %d: %s", par, off, err)
				} else {
					fmt.Println("INFO: produced trade:", producerMessage)
					producedCount++
				}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... closing trade producer...")
				(*tradeProducer).Close()
				tradeChannel <- Trade{}
			}
		}
	}(tradeChannel)

	pricePointChannel := make(chan PricePoint)
	go func(pricePointMessages <-chan PricePoint) {
		for {
			select {
			case pricePoint := <-pricePointMessages:
				producerMessage := sarama.ProducerMessage{Topic: me.pricePointTopic, Value: sarama.StringEncoder(pricePoint.toJSON())}
				par, off, err := (*pricePointProducer).SendMessage(&producerMessage)
				if err != nil {
					fmt.Printf("ERROR: producing price point in partition %d, offset %d: %s", par, off, err)
				} else {
					fmt.Println("INFO: produced price point:", producerMessage)
					producedCount++
				}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... closing price point producer...")
				(*pricePointProducer).Close()
				pricePointChannel <- PricePoint{}
			}
		}
	}(pricePointChannel)

	orderChannel := make(chan []byte)
	go func() {
		for {
			select {
			case msg := <-orderMessages:
				order, err := messageToOrder(msg.Value)
				if err != nil {
					fmt.Println("ERROR: malformed message:", msg.Value)
				} else {
					fmt.Println("DEBUG: received quote:", order)
					consumedCount++
				}
				me.Process(&order, tradeChannel, pricePointChannel)
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				orderChannel <- []byte{}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... Closing quote consummer...")
				orderChannel <- []byte{}
			}
		}
	}()
	<-orderChannel
	fmt.Println("INFO: closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}

func (me *MatchingEngine) Process(inOrder *Order, producerChannel chan<- Trade, pricePointChannel chan<- PricePoint) {
	// var currentBook *map[float64][]*Order
	var oppositeBook *map[float64][]*Order
	// var currentBestPrice Heap
	var oppositeBestPrice Heap
	var action string
	var comparator func(x, y float64) bool

	if inOrder.Quantity > 0 {
		// currentBook = &me.orderBook.PriceToBuyOrders
		oppositeBook = &me.orderBook.PriceToSellOrders
		// currentBestPrice = me.orderBook.BestBid
		oppositeBestPrice = me.orderBook.BestAsk
		action = "buy"
		comparator = func(x, y float64) bool { return x >= y }
	} else {
		// currentBook = &me.orderBook.PriceToSellOrders
		oppositeBook = &me.orderBook.PriceToBuyOrders
		// currentBestPrice = me.orderBook.BestAsk
		oppositeBestPrice = me.orderBook.BestBid
		action = "sell"
		inOrder.Quantity = -inOrder.Quantity
		comparator = func(x, y float64) bool { return x <= y }
	}

	// loop on opposite book
	for inOrder.Quantity > 0 && oppositeBestPrice.Len() > 0 && comparator(inOrder.Price, oppositeBestPrice.Peak().(float64)) {
		oppositeBestPriceQueue := (*oppositeBook)[oppositeBestPrice.Peak().(float64)]
		// loop on nest price queue
		for inOrder.Quantity > 0 && len(oppositeBestPriceQueue) > 0 {
			outOrder := cut(0, &oppositeBestPriceQueue)
			tradeQuantity := math.Min(inOrder.Quantity, outOrder.Quantity)
			price := outOrder.Price
			fmt.Printf(
				"INFO: Executed trade: %s %f @ %f | Left Order ID: %s, Right Order ID: %s | "+
					"Left Order Quantity: %f, Right Order Quantity: %f\n",
				action, tradeQuantity, price, inOrder.OrderID, outOrder.OrderID,
				inOrder.Quantity, outOrder.Quantity,
			)

			inOrder.Quantity -= tradeQuantity
			outOrder.Quantity -= tradeQuantity

			if producerChannel != nil {
				tradeId := uuid.New().String()
				producerChannel <- createTrade(tradeId, inOrder, tradeQuantity, price, action)
				producerChannel <- createTrade(tradeId, outOrder, tradeQuantity, price, action) // TODO: action is opposite for out order
			}
			if pricePointChannel != nil {
				pricePointChannel <- createPricePoint(price)
			}

			if outOrder.Quantity > 0 {
				me.orderBook.AddOrder(outOrder)
			}
		}
		if len(oppositeBestPriceQueue) == 0 {
			bestPrice := oppositeBestPrice.Pop().(float64)
			delete(*oppositeBook, bestPrice)
		}
	}
	if inOrder.Quantity > 0 {
		me.orderBook.AddOrder(inOrder)
	}
}
