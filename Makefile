target: build

# Build all images

build_kafkainit:
	docker build -t local/kafka-init -f src/kafka_init/Dockerfile src/kafka_init/

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src/orderbook/

build_traderpool:
	docker build --no-cache -t local/traderpool -f src/traderpool/Dockerfile src/traderpool/

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update

clear_helm:
	helm repo remove bitnami
	helm repo remove prometheus-community

start_infra:
	kubectl apply -f k8s/namespaces.yaml

start_kafka:
	helm upgrade --install bitnami bitnami/kafka --version 31.0.0 -n orderbook --create-namespace -f helm/kafka/values-local.yaml

forward_kafka:
	kubectl port-forward --namespace orderbook svc/bitnami-kafka-controller-0-external 9094:9094

stop_kafka:
	helm uninstall --ignore-not-found bitnami -n orderbook
	kubectl delete --ignore-not-found pvc data-bitnami-kafka-controller-0 -n orderbook

dev:
	uv pip install -r requirements-dev.txt

test:
	pytest tests/ -vv

build: build_kafkainit build_orderbook build_traderpool

start:start_infra start_kafka
	helm install orderbook ./chart --namespace orderbook

stop:stop_kafka
	helm uninstall --ignore-not-found orderbook --namespace orderbook