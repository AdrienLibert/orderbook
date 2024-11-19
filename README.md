# Order Book

## Dev setup

Setup some local python dev toolkit. Latest stable python on pyenv is 3.12.1.
```
brew install pyenv
pyenv install 3.12.1
pyenv virtualenv 3.12.1 orderbook
pyenv activate orderbook
```

Install local requirements will need you to build our custom library named "drgn".
```
make build_lib
uv pip install -r requirements-dev.txt --find-links $PWD/src/drgn/dist/
```

## Local

### Dependencies

Local mode requires some dependencies. Following commands are for macos, adapt to your OS.
```
brew install docker
brew install helm
brew install kubectl
```

### Build

Build images for order book agent and trader pool agent(s): code is shipped inside images.
```
make build
```


### Helm

Kafka is robust, we will fully leverage helm chart templating: we need to add the helm repo.
```
make helm
```
If other charts are used in the future, they will be added here.

### Start

Start Kafka relied on Helm install. Some custom values are provided in helm-values directory.
```
make start_kafka
make start_kafkainit # will be moved later to start_kafka as part of the start protocol
```

Start agents from custom built images.
```
make start_orderbook
make start_traderpool
```

 ## Kafka

 ### Topic

Orderbook
Order

### Schema

Order :
```
{
  "order_id": { "type": "int", "description": "Unique identifier for the order" },
  "order_type": { "type": "str", "enum": ["buy", "sell"], "description": "Type of order" },
  "price": { "type": "int", "description": "Price of the order" },
  "quantity": { "type": "int", "minimum": 1, "description": "Quantity of the asset" }
}
```
Orderbook :
```
{
  "bids": [
    { "price": { "type": "int" }, "quantity": { "type": "int" } }
  ],
  "asks": [
    { "price": { "type": "int" }, "quantity": { "type": "int" } }
  ]
}
```
