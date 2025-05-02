package main

type Order struct {
	OrderID   string  `json:"order_id"`
	OrderType string  `json:"order_type"`
	Price     float64 `json:"price"`
	Quantity  int64   `json:"quantity"`
	Timestamp int64   `json:"timestamp"`
}

type Trade struct {
	TradeId   string  `json:"trade_id"`
	OrderId   string  `json:"order_id"`
	Quantity  int64   `json:"quantity"`
	Price     float64 `json:"price"`
	Action    string  `json:"action"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
}
type PricePoint struct {
	Price float64 `json:"price"`
}

type Trader struct {
	TradeId string  `json:"trade_id"`
	Price   float64 `json:"price"`
}
