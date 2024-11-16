build: build_kafka build_orderbook build_traderpool

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update

clear_helm:
	helm repo remove bitnami

start_kafka:
	helm upgrade --install bitnami bitnami/kafka --version 31.0.0 -n orderbook --create-namespace -f helm-values/values-local.yaml

stop_kafka:
	helm uninstall bitnami -n orderbook

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src/orderbook/

build_traderpool:
	docker build -t local/traderpool -f src/traderpool/Dockerfile src/traderpool/

make start: start_kafka start_orderbook start traderpool