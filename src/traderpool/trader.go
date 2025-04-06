package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"math/rand"
	"os"
	"os/signal"
)

func (t *Trader) TargetPrices() {
	t.TargetBuy = t.MidPrice.Value + (t.AggressivenessBuy * 0.5)
	t.TargetSell = t.MidPrice.Value - (t.AggressivenessSell * 0.5)
}

func (t *Trader) Trade(orderListChannel chan<- Order) Order {
	t.TargetPrices()

	orderQuantity := int64(rand.Intn(50) + 1)
	orderQuantity *= t.Sign

	var target float64
	if orderQuantity > 0 {
		target = t.TargetBuy
	} else {
		target = t.TargetSell
	}

	order := publishOrder(orderQuantity, target)
	if order.Quantity != 0 {
		select {
		case orderListChannel <- order:
		default:
			fmt.Println("INFO: Order channel full, skipping order:", order)
		}
	} else {
		fmt.Println("INFO: Skipping empty order")
	}
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
			if msg.Quantity == 0 {
				fmt.Println("INFO: Skipping empty order in producer")
				continue
			}
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
								t.Sign = -1 * t.Sign
							}
						default:
							t.AggressivenessSell *= 1.4
							t.AggressivenessBuy *= 1.4
							t.Sign = -1 * t.Sign
						}
					}
				}
			case priceMsg := <-priceConsumer:
				pricePoint, err := convertMessageToPricePoint(priceMsg.Value)
				if err != nil {
					fmt.Printf("ERROR: Failed to parse PricePoint for Trader %s: %s\n", t.TraderId, err)
				} else {
					t.MidPrice.Value = pricePoint.Price
					consumedCount++
					fmt.Printf("INFO: Updated midprice to last price: %.2f\n", t.MidPrice.Value)
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
