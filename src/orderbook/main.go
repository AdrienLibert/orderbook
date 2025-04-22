package main

import (
	"fmt"
	"os"
)

// utils
func cut(i int, orderSlicesRef *[]*Order) (*Order, error) {
	if i < 0 || i >= len(*orderSlicesRef) {
		return nil, fmt.Errorf("index out of bounds: %d", i)
	}
	removedOrder := (*orderSlicesRef)[i]
	*orderSlicesRef = append((*orderSlicesRef)[:i], (*orderSlicesRef)[i+1:]...)
	return removedOrder, nil
}

func getenv(key, fallback string) string {
	// TODO: refactor through a configparser
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	fmt.Println("INFO: tarting orderbook")
	me := NewMatchingEngine(
		NewKafkaClient(),
		NewOrderBook(),
	)
	me.Start()
}
