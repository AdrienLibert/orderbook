package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	fmt.Println("INFO: starting traders")

	rand.Seed(time.Now().UnixNano())

	numTradersStr := getenv("NUM_TRADERS", "10")
	numTraders, err := strconv.Atoi(numTradersStr)
	if err != nil {
		fmt.Printf("ERROR: NUM_TRADERS invalid (%s)\n", numTradersStr)
		os.Exit(1)
	}

	spreadStr := getenv("OB__ORDERBOOK__SPREAD", "1")
	spread, err := strconv.ParseFloat(spreadStr, 64)
	if err != nil {
		fmt.Printf("ERROR: SPREAD invalid (%s)\n", spreadStr)
		os.Exit(1)
	}

	kc := NewKafkaClient()
	master := kc.GetConsumer()
	priceConsumer, _ := kc.Assign(*master, "order.last_price.topic")
	var initialMidPrice float64
	priceMsg := <-priceConsumer
	pricePoint, err := messageToPricePoint(priceMsg.Value)
	if err != nil {
		fmt.Printf("ERROR: Failed to parse initial PricePoint: %v\n", err)
		initialMidPrice = 47.0
	} else {
		initialMidPrice = pricePoint.Price
		fmt.Printf("INFO: Initial MidPrice set to: %.2f\n", initialMidPrice)
	}
	sharedMidPrice := createMidPrice(initialMidPrice, spread)

	var wg sync.WaitGroup

	for i := 0; i < numTraders; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			t := NewTrader(i, sharedMidPrice, kc)
			t.Start()
		}(i)
	}

	wg.Wait()
}
