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

	numTradersStr := os.Getenv("NUM_TRADERS")
	if numTradersStr == "" {
		numTradersStr = "10"
	}

	numTraders, err := strconv.Atoi(numTradersStr)
	if err != nil {
		fmt.Printf("ERROR: NUM_TRADERS invalid (%s)", numTradersStr)
		os.Exit(1)
	}

	spreadStr := os.Getenv("SPREAD")
	if spreadStr == "" {
		spreadStr = "1"
	}

	spread, err := strconv.ParseFloat(spreadStr, 64)
	if err != nil {
		fmt.Printf("ERROR: SPREAD invalid (%s)", spreadStr)
		os.Exit(1)
	}

	midPriceStr := os.Getenv("MID_PRICE")
	if midPriceStr == "" {
		midPriceStr = "45"
	}

	midPrice, err := strconv.ParseFloat(midPriceStr, 64)
	if err != nil {
		fmt.Printf("ERROR: MID_PRICE invalid (%s)", midPriceStr)
		os.Exit(1)
	}

	sharedMidPrice := createMidPrice(midPrice, spread)

	var j float64 = -1
	var wg sync.WaitGroup

	for i := 0; i < numTraders; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			j = j * -1
			t := NewTrader(
				strconv.Itoa(i),
				float64(j)*float64(rand.Intn(11)+5),
				sharedMidPrice,
				NewKafkaClient(),
			)
			t.Start()
		}(i)
	}

	wg.Wait()
}
