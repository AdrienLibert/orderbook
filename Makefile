# Build all images
build: build_drgn build_kafkainit build_orderbook build_traderpool build_influxdb

build_drgn:
	cd src/drgn && \
	uv build

build_kafkainit:
	docker build -t local/kafka-init -f src/kafka_init/Dockerfile src

build_orderbook:
	docker build -t local/orderbook -f src/orderbook/Dockerfile src

build_traderpool:
	docker build --no-cache -t local/traderpool -f src/traderpool/Dockerfile src

build_flink_image:
	docker build -t local/flink -f src/flink/Dockerfile.simple src/flink/

build_flink_job:
	docker build -t local/flink-jobs -f src/flink/Dockerfile src/flink/

build_kustomize:
	docker pull registry.k8s.io/kustomize/kustomize:v5.6.0
	docker run --rm registry.k8s.io/kustomize/kustomize:v5.6.0

build_influxdb:
	helm install my-influxdb bitnami/influxdb -n analytics -f helm/influxdb/values.yaml

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
	kubectl apply -f k8s/namespaces.yaml
	kubectl apply -f k8s/traderpool/

stop_traderpool:
	kubectl delete -f k8s/traderpool/ --ignore-not-found

start_flink_on_k8s: start_infra
	helm install cert-manager jetstack/cert-manager --namespace analytics --version v1.17.1 -f helm/certmanager/values-local.yaml
	helm install flink-kubernetes-operator flink-operator-repo/flink-kubernetes-operator --namespace analytics --version 1.10.0 -f helm/flink-kubernetes-operator/values-local.yaml

stop_flink_on_k8s:
	helm uninstall --ignore-not-found cert-manager -n analytics
	helm uninstall --ignore-not-found flink-kubernetes-operator -n analytics
	kubectl delete crd issuers.cert-manager.io clusterissuers.cert-manager.io certificates.cert-manager.io certificaterequests.cert-manager.io orders.acme.cert-manager.io challenges.acme.cert-manager.io  --ignore-not-found
	kubectl delete crd flinkclusters.flinkoperator.k8s.io --ignore-not-found
	kubectl delete secret webhook-server-cert -n flink --ignore-not-found
	kubectl delete secret cert-manager-webhook-ca -n flink --ignore-not-found
	kubectl delete job cert-manager-startupapicheck -n flink --ignore-not-found

start_flink_example:
	kubectl apply -f k8s/flink/example-deployment.yaml

stop_flink_example:
	kubectl delete -f k8s/flink/example-deployment.yaml --ignore-not-found

start_flink_custom_image:
	kubectl apply -f k8s/flink/example-custom-image.yaml

stop_flink_custom_image:
	kubectl delete -f k8s/flink/example-custom-image.yaml --ignore-not-found

start_flink_custom_job:
	kubectl apply -f k8s/flink/example-custom-job.yaml

stop_flink_custom_job:
	kubectl delete -f k8s/flink/example-custom-job.yaml --ignore-not-found

start_flink_candle_job:
	kubectl apply -f k8s/flink/candle-stick-job.yaml

stop_flink_candle_job:
	kubectl delete -f k8s/flink/candle-stick-job.yaml --ignore-not-found

start_grafana: build_kustomize
	kustomize build grafana-dashboard | kubectl apply -n monitoring -f -
	helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack -n monitoring -f helm/grafana/values.yaml
	kubectl apply -f k8s/monitoring/ -n monitoring
	kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring

stop_grafana: build_kustomize
	kustomize build grafana-dashboard | kubectl delete -n monitoring -f -
	helm uninstall --ignore-not-found kube-prometheus-stack -n monitoring
	kubectl delete --ignore-not-found pvc kube-prometheus-stack-grafana -n monitoring
	kubectl delete --ignore-not-found svc kafka-exporter -n monitoring
	kubectl delete --ignore-not-found deployment kafka-exporter -n monitoring
	kubectl delete --ignore-not-found svc node-exporter -n monitoring

start_influxdb:
	kubectl port-forward svc/my-influxdb 8086:8086 -n analytics

stop_influxdb:
	helm uninstall my-influxdb -n analytics

start: start_kafka start_orderbook start_traderpool start_influxdb

stop: stop_kafka stop_orderbook stop_traderpool stop_kafkainit stop_flink_on_k8s stop_influxdb

dev:
	uv pip install -r requirements-dev.txt --find-links $$PWD/src/drgn/dist/

test:
	pytest tests/ -vv