# Build all images
build: build_drgn build_kafkainit build_orderbook build_traderpool

build_drgn:
	cd src/drgn && \
	uv build

build_kafkainit:
	docker build -t local/kafka-init -f src/kafka_init/Dockerfile src

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src

build_traderpool:
	docker build --no-cache -t local/traderpool -f src/traderpool/Dockerfile src

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update

clear_helm:
	helm repo remove bitnami
	helm repo remove grafana

start_kafka:
	helm upgrade --install bitnami bitnami/kafka --version 31.0.0 -n orderbook --create-namespace -f helm-values/values-local.yaml
	kubectl apply -f k8s/kafka_init/

forward_kafka:
	kubectl port-forward --namespace orderbook svc/bitnami-kafka-controller-0-external 9094:9094

stop_kafka:
	helm uninstall --ignore-not-found bitnami -n orderbook
	kubectl delete --ignore-not-found -f k8s/kafka_init/
	kubectl delete --ignore-not-found pvc data-bitnami-kafka-controller-0 -n orderbook

start_kafkainit:
	kubectl apply -f k8s/kafka_init/

stop_kafkainit:
	kubectl delete -f k8s/kafka_init/ --ignore-not-found

start_orderbook:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/orderbook/

stop_orderbook:
	kubectl delete -f k8s/orderbook/ --ignore-not-found

start_traderpool:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/traderpool/

stop_traderpool:
	kubectl delete -f k8s/traderpool/ --ignore-not-found

build_grafana:
	helm install grafana grafana/grafana -f k8s/monitoring/values.yaml
	kubectl apply -f k8s/monitoring/kubelet_api.yaml
	TOKEN=$$(kubectl get secret prom-api-user-secret -o jsonpath='{.data.token}' | base64 --decode); \


start_grafana:
	 kubectl apply -f k8s/monitoring/prometheus-deployment.yaml
	 export POD_NAME=$$(kubectl get pods -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=grafana" -o jsonpath="{.items[0].metadata.name}"); \
	 kubectl port-forward $$POD_NAME 3000

stop_grafana:
	helm uninstall --ignore-not-found grafana
	kubectl delete --ignore-not-found pvc grafana
	kubectl delete deployment prometheus

make start: start_kafka start_orderbook start_traderpool

make stop: stop_kafka stop_orderbook stop_traderpool stop_kafkainit

make dev: 
	uv pip install -r requirements-dev.txt --find-links $$PWD/src/drgn/dist/