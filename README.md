# Order Book

Orderbook is a simple implementation of orderbook in python which simulates trader bot swamps trading and a level 1 orderbook (for now).

## Technologies

The whole system is built with following technologies:
- Kafka
- Python
- Docker
- Kubernetes
- Helm
- Pyspark
- minio

### Architecture
![Alt text](_img/architecture.png?raw=true "Architecture Diagram")

The architecture choices are driven by:
- Kafka as efficient and scalable message broker
- Order book in python for matching of orders.
- PySpark for kafka data persistence
- Traders in python for orders simulation

### Message definitions

Orders:
```json
{
    "order_id": "<uuid>",       
    "order_type": "limit",       
    "price": <price>,           
    "quantity": <quantity>,   
    "time": <timestamp>       
}
```
- **order_id**: A unique identifier for the order
- **order_type**: The type of order being placed ("limit", "market")
- **price**: The price at which the order is placed
- **quantity**: The quantity of items. A positive value indicates a "buy" order, while a negative value indicates a "sell" order.
- **time**: The timestamp when the order is created

Trades:
```json
{
    "left_order_id": "<order_id>",     
    "right_order_id": "<order_id>",    
    "quantity": <quantity>,         
    "price": <price>,           
    "action": "<action>", 
    "status": "<status>"              
}
```
- **left_order_id**: The ID of the order on the left side of the trade 
- **right_order_id**: The ID of the order on the right side of the trade
- **quantity**: The number of items traded between the two orders
- **price**: The price at which the trade was executed
- **action**: Specifies whether the trade was a "Buy" or "Sell" action
- **status**: The status of the trade

Prices:
```json
{
    "mid_price": <mid_price>
}
```
- **mid_price**: The most recent mid-price, which is typically the average of the highest bid and lowest ask price for an asset


## Sequence Diagram

TODO

## Getting started

### Local Cluster

#### Dependencies

The ocal cluster requires some dependencies. Following commands are for macos, adapt to your OS.
```
brew install docker
brew install helm
brew install kubectl
```

#### Build

This step builds all images for the various services of the order book stack which for now contains:
- kafka
- kafka_init
- orderbook service
- trader pool service

```
make build
```

#### Helm

This step leverages helm for kafka stack component
```
make helm
```
If other charts are used in the future, they will be added here.

#### Start

Start Kafka relied on Helm install. Some custom values are provided in helm-values directory.
```
make start_kafka
```

Start agents from custom built images.
```
make start_orderbook
make start_traderpool
```

#### Forward
We use a custom NodePort svc for external connection. This is using the 9094 and being relayed on localhost:9094:
```
kubectl port-forward -n orderbook svc/bitnami-kafka-controller-0-external 9094:9094
```

### Dev mode

Setup some local python dev toolkit. Latest stable python on pyenv is 3.12.1.
```
brew install pyenv
pyenv install 3.12.1
pyenv virtualenv 3.12.1 orderbook
pyenv activate orderbook
```

Install local requirements will need you to build our custom library named "drgn".
```
pip install uv==0.5.2
make build_drgn
uv pip install -r requirements-dev.txt --find-links $PWD/src/drgn/dist/
```