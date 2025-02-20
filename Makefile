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

build_flink:
	docker build -t local/flink -f src/flink/Dockerfile src/flink/

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add jetstack https://charts.jetstack.io
	helm repo add flink-operator-repo https://downloads.apache.org/flink/flink-kubernetes-operator-1.10.0/
	helm repo update

clear_helm:
	helm repo remove bitnami
	helm repo remove jetstack

start_infra:
	kubectl apply -f k8s/namespaces.yaml

start_kafka:
	helm upgrade --install bitnami bitnami/kafka --version 31.0.0 -n orderbook --create-namespace -f helm/kafka/values-local.yaml
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
	kubectl apply -f k8s/orderbook/

stop_orderbook:
	kubectl delete -f k8s/orderbook/ --ignore-not-found

start_traderpool:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/traderpool/

stop_traderpool:
	kubectl delete -f k8s/traderpool/ --ignore-not-found

start_flink_on_k8s: 
	start_infra
	helm install cert-manager jetstack/cert-manager --namespace flink --create-namespace --version v1.17.1 -f helm/flink-kubernetes-operator/values-local.yaml
	helm install flink-kubernetes-operator flink-operator-repo/flink-kubernetes-operator --namespace flink --create-namespace -f helm/flink-kubernetes-operator/values-local.yaml

stop_flink_on_k8s:
	helm delete --ignore-not-found cert-manager -n flink
	helm delete --ignore-not-found flink-kubernetes-operator -n flink
	kubectl delete crd issuers.cert-manager.io clusterissuers.cert-manager.io certificates.cert-manager.io certificaterequests.cert-manager.io orders.acme.cert-manager.io challenges.acme.cert-manager.io  --ignore-not-found
	kubectl delete crd flinkclusters.flinkoperator.k8s.io --ignore-not-found
	kubectl delete secret webhook-server-cert -n flink --ignore-not-found
	kubectl delete secret cert-manager-webhook-ca -n flink --ignore-not-found
	kubectl delete job cert-manager-startupapicheck -n flink --ignore-not-found
	kubectl delete namespace flink --ignore-not-found

start_flink_example:
	kubectl apply -f k8s/flink/example-deployment.yaml
	
start: start_kafka start_orderbook start_traderpool

stop: stop_kafka stop_orderbook stop_traderpool stop_kafkainit stop_flink_on_k8s

dev: 
	uv pip install -r requirements-dev.txt --find-links $$PWD/src/drgn/dist/

test:
	pytest tests/ -vv