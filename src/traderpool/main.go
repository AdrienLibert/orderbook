package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	fmt.Printf("NUM_TRADERS %d\n", numTraders)

	kc := NewKafkaClient()
	kc.Start(numTraders)
}
