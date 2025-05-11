package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"os"
	"os/signal"
	"time"
)

func Start(numTraders int, kc *KafkaClient) {
	orderProducer := kc.GetProducer()
	if orderProducer == nil {
		fmt.Println("ERROR: Kafka producer is nil! Exiting.")
		return
	}

	master := kc.GetConsumer()
	consumer, errors := kc.Assign(*master, kc.tradeTopic)
	priceConsumer, errors := kc.Assign(*master, kc.pricePointTopic)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumedCount := 0
	producedCount := 0

	orderListChannel := make(chan Order, 10)
	priceListChannel := make(chan Trader, 10)

	go func(orderMessage <-chan Order) {
		for msg := range orderMessage {
			if msg.Quantity == 0 {
				fmt.Println("INFO: Skipping empty order in producer")
				continue
			}
			producerMessage := sarama.ProducerMessage{
				Topic: kc.quoteTopic,
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

	consumeChannel := make(chan struct{})
	go func(priceChannel chan<- Trader) {
		for {
			select {
			case msg := <-consumer:
				trade, err := convertMessageToTrade(msg.Value)
				if err != nil {
					return
				}
				consumedCount++
				priceChannel <- Trader{Price: trade.Price, TradeId: trade.TradeId}
				fmt.Printf("INFO: Stored price in priceListChannel: %.2f\n", trade.Price)
			case consumerError := <-errors:
				consumedCount++
				fmt.Println("ERROR: received trade consumerError:", consumerError.Err)
				consumeChannel <- struct{}{}

			case <-time.After(2 * time.Second):
				fmt.Println("INFO: No messages consumed, requesting mid price")
				for i := 0; i < numTraders; i++ {
					priceChannel <- Trader{Price: RequestMidPrice(priceConsumer), TradeId: fmt.Sprintf("Trader-%d", i+1)}
				}
			}
		}
	}(priceListChannel)
	for i := 0; i < numTraders; i++ {
		traderID := fmt.Sprintf("Trader-%d", i+1)
		StartTrader(traderID, priceListChannel, orderListChannel)
	}

	<-consumeChannel
}
