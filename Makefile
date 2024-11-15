build: build_kafka build_orderbook build_traderpool

build_kafka:
	echo "kafka"

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src/orderbook/

build_traderpool:
	docker build -t local/traderpool -f src/traderpool/Dockerfile src/traderpool/