# Order Book

## Dev setup

Setup some local python dev toolkit. Latest stable python on pyenv is 3.12.1

```
brew install pyenv
pyenv install 3.12.1
pyenv virtualenv 3.12.1 orderbook
pyenv activate orderbook
```

## Local

### Dependencies

K8S local needs some dependencies
```
brew install docker
brew install helm
```

### Build

Build images
```
make build
```

Start Kafka
```
make start_kafka
```

Start agents
```
make start_orderbook
make start_traderpool
```