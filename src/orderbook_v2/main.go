package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"

	"github.com/IBM/sarama"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp float64 `json:"timestamp"`
}

type KafkaClient struct {
	config   *sarama.Config
	consumer *sarama.Consumer
	brokers  []string
}

func NewKafkaClient() *KafkaClient {
	kc := new(KafkaClient)
	kc.brokers = []string{"localhost:9094"}
	kc.config = sarama.NewConfig()
	kc.config.ClientID = "go-orderbook-consumer"
	kc.config.Consumer.Return.Errors = true
	kc.config.Net.SASL.Enable = false
	kc.config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	kc.config.Consumer.Offsets.Initial = sarama.OffsetNewest
	return kc
}

func (kc *KafkaClient) GetConsumer() *sarama.Consumer {
	if kc.consumer != nil {
		return kc.consumer
	}
	consumer, err := sarama.NewConsumer(kc.brokers, kc.config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	return &consumer
}

func (kc *KafkaClient) Assign(master sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := master.Partitions(topic)
	// debuging here
	topics, err := master.Topics()
	if err != nil {
		panic(err)
	}
	fmt.Println("Topics: ", topics)
	consumer, err := master.ConsumePartition(
		topic,
		partitions[0], // only first partition for now
		sarama.OffsetOldest,
	)

	if err != nil {
		fmt.Printf("Topic %v Partitions %v", topic, partitions)
		panic(err)
	}
	fmt.Println("Start consuming topic: ", topic)

	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				errors <- consumerError
				fmt.Println("consumerError: ", consumerError.Err)
			case msg := <-consumer.Messages():
				consumers <- msg
			}
		}
	}(topic, consumer)

	return consumers, errors
}

type Orderbook struct {
	bid         []*Order
	ask         []*Order
	kafkaClient *KafkaClient

	_quoteTopic string
}

func cut(i int, xs *[]*Order) *Order {
	y := (*xs)[i]
	*xs = append((*xs)[:i], (*xs)[i+1:]...)
	return y
}

func (o *Orderbook) Process(inOrder *Order) {
	var in, out *[]*Order
	var action string
	var comparator func(x, y float64) bool

	if inOrder.Quantity > 0 {
		in = &o.bid
		out = &o.ask
		action = "buy"
		comparator = func(x, y float64) bool { return x >= y }
	} else {
		in = &o.ask
		out = &o.bid
		action = "sell"
		inOrder.Quantity = -inOrder.Quantity
		comparator = func(x, y float64) bool { return x <= y }
	}

	for inOrder.Quantity > 0 && len(*out) > 0 && comparator(inOrder.Price, (*out)[0].Price) {
		outOrder := cut(0, out)
		fmt.Println(action, inOrder, outOrder)

		tradeQuantity := math.Min(inOrder.Quantity, outOrder.Quantity)
		fmt.Printf(
			"Executed trade: %s %f @ %f | Left Order ID: %s, Right Order ID: %s | "+
				"Left Order Quantity: %f, Right Order Quantity: %f\n",
			action, tradeQuantity, outOrder.Price, inOrder.OrderID, outOrder.OrderID,
			inOrder.Quantity, outOrder.Quantity,
		)

		inOrder.Quantity -= tradeQuantity
		outOrder.Quantity -= tradeQuantity

		// TODO: publish orderbook change

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

func (o *Orderbook) Start() {
	master := o.kafkaClient.GetConsumer()
	consumer, errors := o.kafkaClient.Assign(*master, o._quoteTopic)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	msgCount := 0
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				order, err := convertMessageToOrder(msg.Value)
				if err != nil {
					handleError(err)
				}
				o.Process(&order)
			case consumerError := <-errors:
				msgCount++
				fmt.Println("Received consumerError:", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
	fmt.Println("Closing... processed", msgCount, "messages")
}

func NewOrderBook(bid []*Order, ask []*Order, kafkaClient *KafkaClient) *Orderbook {
	o := new(Orderbook)
	o.bid = bid
	o.ask = ask
	o.kafkaClient = kafkaClient
	o._quoteTopic = "orders.topic"
	return o
}

func handleError(err error) {
	fmt.Println("invalid message consummed:", err)
}

func convertMessageToOrder(messageValue []byte) (Order, error) {
	var order = &Order{}
	if err := json.Unmarshal(messageValue, order); err != nil {
		return Order{}, err
	}

	return *order, nil
}

func main() {
	fmt.Println("Starting orderbook")
	o := NewOrderBook(
		[]*Order{},
		[]*Order{},
		NewKafkaClient(),
	)
	o.Start()
}
