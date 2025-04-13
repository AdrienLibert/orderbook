# Order Book

## Structures
### Orderbook

The Order represents a simple limit order.
```
type Order struct {
	OrderID   string
	OrderType string
	Price     float64
	Quantity  float64
	Timestamp int64
}
```

The Orderbook is the actual data structure holding the orders ready to be filled. Main operations in an orderbook are
- place order
- get volume at price
- delete (optional)

```
type Orderbook struct {
	BestBid       *MaxHeap
	BestAsk       *MinHeap
	PriceToVolume map[float64]float64
	PriceToOrders map[float64]Order
}
```

## Local Cluster

### Dependencies

The ocal cluster requires some dependencies. Following commands are for macos, adapt to your OS.
```
brew install docker
brew install helm
brew install kubectl
```

### Build

This step builds all images for the various services of the order book stack which for now contains:
- kafka_init
- orderbook service
- trader pool service

```
make build
```

### Helm

This step leverages helm for kafka stack component
```
make helm
```

### Start

Star the whole infrastructure and run the services
```
make start_deps
make start
```