package main

import (
	"fmt"
	"os"
)

// utils
func cut(i int, xs *[]*Order) *Order {
	y := (*xs)[i]
	*xs = append((*xs)[:i], (*xs)[i+1:]...)
	return y
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
	fmt.Println("INFO: starting orderbook")
	me := NewMatchingEngine(
		NewKafkaClient(),
		NewOrderBook(),
	)
	me.Start()
}
