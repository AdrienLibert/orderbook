# Build all images

alias kustomize docker run --rm registry.k8s.io/kustomize/kustomize:v5.6.0

build: build_drgn build_kafkainit build_orderbook build_traderpool build_flink build_kustomize

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
	docker build -t local/flink-jobs -f src/flink/Dockerfile src/flink/

build_spark:
	docker build -t local/spark:3.5.5 -f src/spark/Dockerfile src

build_kustomize:
	docker pull registry.k8s.io/kustomize/kustomize:v5.6.0

helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add jetstack https://charts.jetstack.io
	helm repo add flink-operator-repo https://downloads.apache.org/flink/flink-kubernetes-operator-1.10.0/
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add spark-operator https://kubeflow.github.io/spark-operator
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

start_orderbook:
	helm install orderbook ./orderbook_chart

stop_orderbook:
	helm uninstall orderbook

start_flink_on_k8s: start_infra
	helm install flink-kubernetes-operator flink-operator-repo/flink-kubernetes-operator --namespace analytics --version 1.10.0 -f helm/flink-k8s-operator/values-local.yaml

stop_flink_on_k8s:
	helm uninstall --ignore-not-found flink-kubernetes-operator -n analytics
	kubectl delete crd flinkclusters.flinkoperator.k8s.io --ignore-not-found

start_flink:
	kubectl apply -f k8s/flink/candle-stick-job.yaml

stop_flink:
	kubectl delete -f k8s/flink/candle-stick-job.yaml --ignore-not-found

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

start_db: start_infra
	kubectl create secret generic postgres --from-literal=password=postgres --from-literal=postgres-password=postgres --dry-run -o yaml | kubectl apply -f -
	helm upgrade --install postgres bitnami/postgresql --version 16.5.6 -n analytics --create-namespace -f helm/postgres/values-local.yaml

stop_db:
	helm uninstall --ignore-not-found postgres -n analytics
	kubectl delete --ignore-not-found pvc data-postgres-postgresql-0 -n analytics
	kubectl delete --ignore-not-found secret postgres

forward_db:
	kubectl port-forward svc/postgres-postgresql 5432:5432 -n analytics

start_spark: start_infra
	helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace --wait
	kubectl apply -f k8s/spark/

stop_spark:
	helm uninstall --ignore-not-found spark-operator -n spark-operator

start: start_kafka start_orderbook

stop: stop_kafka stop_orderbook

dev:
	uv pip install -r requirements-dev.txt --find-links $$PWD/src/drgn/dist/

test:
	pytest tests/ -vv