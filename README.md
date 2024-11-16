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
```

Start agents from custom built images.
```
make start_orderbook
make start_traderpool
```