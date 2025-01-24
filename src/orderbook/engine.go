package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
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
	me.tradeTopic = "order.status.topic"
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
					handleError(err)
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
	var currentBook *map[float64][]*Order
	var oppositeBook *map[float64][]*Order
	var currentBestPriceHeap Heap
	var oppositeBestPriceHeap Heap
	var oppositeBestPrice float64
	var action string
	var comparator func(x, y float64) bool

	if inOrder.Quantity > 0 {
		currentBook = &me.orderBook.PriceToBuyOrders
		oppositeBook = &me.orderBook.PriceToSellOrders
		currentBestPriceHeap = me.orderBook.BestBid
		oppositeBestPriceHeap = me.orderBook.BestAsk
		oppositeBestPrice = oppositeBestPriceHeap.Pop().(float64)
		action = "buy"
		comparator = func(x, y float64) bool { return x >= y }
	} else {
		currentBook = &me.orderBook.PriceToSellOrders
		oppositeBook = &me.orderBook.PriceToBuyOrders
		currentBestPriceHeap = me.orderBook.BestAsk
		oppositeBestPriceHeap = me.orderBook.BestBid
		oppositeBestPrice = oppositeBestPriceHeap.Pop().(float64)
		action = "sell"
		inOrder.Quantity = -inOrder.Quantity
		comparator = func(x, y float64) bool { return x <= y }
	}
	inPrice := inOrder.Price

	for inOrder.Quantity > 0 && len(*oppositeBook) > 0 && comparator(inOrder.Price, oppositeBestPrice) {
		// do something
		fmt.Println(action)
	}
	if inOrder.Quantity > 0 {
		currentBestPriceHeap.Push(inPrice)
		(*currentBook)[inPrice] = append((*currentBook)[inPrice], inOrder)
	}
	//
	//for inOrder.Quantity > 0 && len(*oppositeBook) > 0 && comparator(inOrder.Price, (*oppositeBook)[0].Price) {
	//	outOrder := cut(0, oppositeBook)
	//	fmt.Println(action, inOrder, outOrder)
	//
	//	tradeQuantity := math.Min(inOrder.Quantity, outOrder.Quantity)
	//	price := outOrder.Price
	//	fmt.Printf(
	//		"INFO: Executed trade: %s %f @ %f | Left Order ID: %s, Right Order ID: %s | "+
	//			"Left Order Quantity: %f, Right Order Quantity: %f\n",
	//		action, tradeQuantity, price, inOrder.OrderID, outOrder.OrderID,
	//		inOrder.Quantity, outOrder.Quantity,
	//	)
	//
	//	inOrder.Quantity -= tradeQuantity
	//	outOrder.Quantity -= tradeQuantity
	//
	//	producerChannel <- publishTrade(inOrder, tradeQuantity, price, action)
	//	producerChannel <- publishTrade(outOrder, tradeQuantity, price, action)
	//	pricePointChannel <- publishPricePoint(price)
	//
	//	if outOrder.Quantity > 0 {
	//		*out = append(*out, outOrder)
	//		sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
	//	}
	//}
	//
	//if inOrder.Quantity > 0 {
	//	*in = append(*in, inOrder)
	//	sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
	//}
}
