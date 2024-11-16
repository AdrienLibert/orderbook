build: build_kafka build_orderbook build_traderpool

helm:
	sudo apt-get update && \
	sudo apt-get install -y helm && \
	helm repo add bitnami https://charts.bitnami.com/bitnami && \
	helm install kafka oci://registry-1.docker.io/bitnamicharts/kafka

start_kafka:


stop_kafka:

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src/orderbook/

build_traderpool:
	docker build -t local/traderpool -f src/traderpool/Dockerfile src/traderpool/