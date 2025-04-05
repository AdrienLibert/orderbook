package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"math/rand"
	"os"
	"os/signal"
)

func (t *Trader) TargetPrices() {
	bid := t.MidPrice.Value - (t.MidPrice.Spread / 2)
	ask := t.MidPrice.Value + (t.MidPrice.Spread / 2)
	t.TargetBuy = t.MidPrice.Value + (t.AggressivenessBuy * (ask - t.MidPrice.Value))
	t.TargetSell = t.MidPrice.Value - (t.AggressivenessSell * (t.MidPrice.Value - bid))
}

func (t *Trader) Trade(orderListChannel chan<- Order) Order {
	t.TargetPrices()

	baseQuantity := int64(rand.Intn(50) + 1)
	sign := int64(1)
	if rand.Float64() < 0.5 {
		sign = -1.0
	}
	t.Quantity = baseQuantity * sign

	var orderType string
	var target float64

	if t.Quantity > 0 {
		orderType = "buy"
		target = t.TargetBuy
	} else {
		orderType = "sell"
		target = t.TargetSell
	}

	order := publishOrder(t.Quantity, target, orderType)
	orderListChannel <- order
	return order
}

func (t *Trader) Start() {
	orderProducer := t.KafkaClient.GetProducer()
	if orderProducer == nil {
		fmt.Println("ERROR: Kafka producer is nil! Exiting.")
		return
	}

	master := t.KafkaClient.GetConsumer()
	consumer, errors := t.KafkaClient.Assign(*master, t.TradeTopic)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	orderListChannel := make(chan Order, 10)

	go func(orderMessage <-chan Order) {
		for msg := range orderMessage {
			producerMessage := sarama.ProducerMessage{
				Topic: t.QuoteTopic,
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
			}
		}
	}()

	consumeChannel := make(chan struct{})
	priceConsumer, priceErrors := t.KafkaClient.Assign(*master, t.PricePointTopic)
	go func() {
		for {
			select {
			case msg := <-consumer:
				trade, err := convertMessageToTrade(msg.Value)
				if err != nil {
					handleError(err)
				} else {
					if trade.OrderId == t.TraderId {
						consumedCount++
						switch trade.Status {
						case "closed":
							if rand.Float64() < 0.75 {
								t.Quantity = -t.Quantity
							}
						case "partial":
							if rand.Float64() < 0.5 {
								t.Quantity = -t.Quantity
							}
							if rand.Float64() < 0.5 {
								t.AggressivenessSell = t.AggressivenessSell * 1.2
								t.AggressivenessBuy = t.AggressivenessBuy * 1.2
							} else {
								t.AggressivenessSell = t.AggressivenessSell * 0.8
								t.AggressivenessBuy = t.AggressivenessBuy * 0.8
							}
						}
					}
				}
			case priceMsg := <-priceConsumer:
				pricePoint, err := messageToPricePoint(priceMsg.Value)
				if err != nil {
					fmt.Printf("ERROR: Failed to parse PricePoint for Trader %s: %s\n", t.TraderId, err)
				} else {
					t.MidPrice.Value = pricePoint.Price
					consumedCount++
					fmt.Printf("INFO: Update midprice to last price: %.2f\n", t.MidPrice.Value)
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
