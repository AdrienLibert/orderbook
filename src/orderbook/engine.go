package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"

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
	master := me.kafkaClient.GetConsumer()
	consumer, errors := me.kafkaClient.Assign(*master, me.quoteTopic)

	tradeProducer := me.kafkaClient.GetProducer()
	pricePointProducer := me.kafkaClient.GetProducer()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	traderChannel := make(chan Trade)
	go func(tradeMessage <-chan Trade) {
		for {
			select {
			case msg := <-tradeMessage:
				producerMessage := sarama.ProducerMessage{Topic: me.tradeTopic, Value: sarama.StringEncoder(convertTradeToMessage(msg))}
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
				traderChannel <- Trade{}
			}
		}
	}(traderChannel)

	pricePointChannel := make(chan PricePoint)
	go func(pricePointMessage <-chan PricePoint) {
		for {
			select {
			case msg := <-pricePointMessage:
				producerMessage := sarama.ProducerMessage{Topic: me.pricePointTopic, Value: sarama.StringEncoder(convertPricePointToMessage(msg))}
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

	consumeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				order, err := convertMessageToOrder(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					fmt.Println("DEBUG: received quote:", order)
					consumedCount++
				}
				me.Process(&order, traderChannel, pricePointChannel)
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("INFO: interrupt is detected... Closing quote consummer...")
				consumeChannel <- struct{}{}
			}
		}
	}()
	<-consumeChannel
	fmt.Println("INFO: closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}

func (o *Orderbook) Process(inOrder *Order, producerChannel chan<- Trade, pricePointChannel chan<- PricePoint) {
	var currentBook *map[float64]Order
	var oppositeBook *map[float64]Order
	var action string
	var comparator func(x, y float64) bool

	if inOrder.Quantity > 0 {
		currentBook = &o.PriceToBuyOrders
		oppositeBook = &o.PriceToSellOrders
		action = "buy"
		comparator = func(x, y float64) bool { return x >= y }
	} else {
		currentBook = &o.PriceToSellOrders
		oppositeBook = &o.PriceToBuyOrders
		action = "sell"
		inOrder.Quantity = -inOrder.Quantity
		comparator = func(x, y float64) bool { return x <= y }
	}

	for inOrder.Quantity > 0 && len(*oppositeBook) > 0 && comparator(inOrder.Price, (*oppositeBook)[0].Price) {
		outOrder := cut(0, oppositeBook)
		fmt.Println(action, inOrder, outOrder)

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

		producerChannel <- publishTrade(inOrder, tradeQuantity, price, action)
		producerChannel <- publishTrade(outOrder, tradeQuantity, price, action)
		pricePointChannel <- publishPricePoint(price)

		if outOrder.Quantity > 0 {
			*out = append(*out, outOrder)
			sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
		}
	}

	if inOrder.Quantity > 0 {
		*in = append(*in, inOrder)
		sort.Slice(*in, func(i, j int) bool { return (*in)[i].Price > (*in)[j].Price })
	}
}
