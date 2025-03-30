package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"
)

func (t *Trader) TargetPrices() {
	bid := t.midPrice.Value - (t.midPrice.Spread / 2)
	ask := t.midPrice.Value + (t.midPrice.Spread / 2)
	t.targetBuy = t.midPrice.Value + (t.aggressivenessBuy * (ask - t.midPrice.Value))
	t.targetSell = t.midPrice.Value - (t.aggressivenessSell * (t.midPrice.Value - bid))
}

func (t *Trader) Trade(orderListChannel chan<- Order) Order {
	t.TargetPrices()
	var orderType string
	var target float64

	if t.quantity > 0 {
		orderType = "buy"
		target = t.targetBuy
	} else {
		orderType = "sell"
		target = t.targetSell
	}

	order := publishOrder(t.traderId, t.quantity, target, orderType)
	orderListChannel <- order
	return order
}

func (t *Trader) Start() {
	orderProducer := t.kafkaClient.GetProducer()
	if orderProducer == nil {
		fmt.Println("ERROR: Kafka producer is nil! Exiting.")
		return
	}

	master := t.kafkaClient.GetConsumer()
	consumer, errors := t.kafkaClient.Assign(*master, t._tradeTopic)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	orderListChannel := make(chan Order, 10)

	go func(orderMessage <-chan Order) {
		for msg := range orderMessage {
			producerMessage := sarama.ProducerMessage{
				Topic: t._quoteTopic,
				Value: sarama.StringEncoder(convertOrderToMessage(msg)),
			}
			par, off, err := (*orderProducer).SendMessage(&producerMessage)
			if err != nil {
				fmt.Printf("ERROR: producing order in partition %d, offset %d: %s\n", par, off, err)
			} else {
				fmt.Println("INFO: produced order:", msg)
				producedCount++
			}
		}
	}(orderListChannel)

	go func() {
		for {
			select {
			case <-signals:
				fmt.Println("INFO: Interrupt detected... Stopping traders.")
				close(orderListChannel)
				return
			default:
				t.Trade(orderListChannel)
				time.Sleep(time.Second * 2)
			}
		}
	}()

	consumeChannel := make(chan struct{})
	priceConsumer, priceErrors := t.kafkaClient.Assign(*master, t._pricePointTopic)
	go func() {
		for {
			select {
			case msg := <-consumer:
				trade, err := convertMessageToTrade(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					if (trade.OrderId == t.traderId) && (trade.Status != "closed") {
						consumedCount++
					}
				}
			case priceMsg := <-priceConsumer:
				pricePoint, err := messageToPricePoint(priceMsg.Value)
				if err != nil {
					fmt.Printf("ERROR: Failed to parse PricePoint for Trader %s: %s\n", t.traderId, err)
				} else {
					t.midPrice.Value = pricePoint.Price
					consumedCount++
					fmt.Printf("INFO: Trader %s updated midprice to last price: %.2f\n", t.traderId, t.midPrice.Value)
				}
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received trade consumerError:", consumerError.Err)
				consumeChannel <- struct{}{}
			case priceError := <-priceErrors:
				consumedCount++
				fmt.Println("ERROR: received pricePoint consumerError:", priceError.Err)
				consumeChannel <- struct{}{}
			case <-signals:
				fmt.Println("INFO: Interrupt detected... Closing trade consumer...")
				consumeChannel <- struct{}{}
			}
		}
	}()

	<-consumeChannel
	fmt.Println("INFO: Closing... processed", consumedCount, "messages and produced", producedCount, "messages")
}
