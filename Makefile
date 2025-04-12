target: build

# Build all images
alias kustomize docker run --rm registry.k8s.io/kustomize/kustomize:v5.6.0

build_kafkainit:
	docker build -t local/kafka-init -f src/kafka_init/Dockerfile src/kafka_init/

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src/orderbook/

build_traderpool:
	docker build --no-cache -t local/traderpool -f src/traderpool/Dockerfile src/traderpool/

build_kustomize:
	docker pull registry.k8s.io/kustomize/kustomize:v5.6.0

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add jetstack https://charts.jetstack.io
	helm repo add flink-operator-repo https://downloads.apache.org/flink/flink-kubernetes-operator-1.10.0/
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update

clear_helm:
	helm repo remove bitnami
	helm repo remove jetstack
	helm repo remove flink-operator-repo
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

start_grafana: build_kustomize
	kustomize build src/grafana-dashboard | kubectl apply -n monitoring -f -
	helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack -n monitoring -f helm/grafana/values-local.yaml
	kubectl apply -f k8s/grafana/ -n monitoring

stop_grafana: build_kustomize
	kustomize build src/grafana-dashboard | kubectl delete -n monitoring -f -
	helm uninstall --ignore-not-found kube-prometheus-stack -n monitoring
	kubectl delete --ignore-not-found pvc kube-prometheus-stack-grafana -n monitoring
	kubectl delete --ignore-not-found svc kafka-exporter -n monitoring
	kubectl delete --ignore-not-found deployment kafka-exporter -n monitoring
	kubectl delete --ignore-not-found svc node-exporter -n monitoring
	kubectl delete --ignore-not-found service grafana-service -n monitoring

dev:
	uv pip install -r requirements-dev.txt

test:
	pytest tests/ -vv

build: build_drgn build_kafkainit build_orderbook build_traderpool

start:
	helm install orderbook ./chart --namespace orderbook

stop:
	helm uninstall --ignore-not-found orderbook --namespace orderbook