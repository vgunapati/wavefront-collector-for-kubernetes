source ./k8s-utils.sh
wait_for_cluster_ready

echo "deploying prometheus endpoint"

kubectl apply -f prom-example.yaml

echo "deploying memcached"

helm repo add bitnami https://charts.bitnami.com/bitnami || true
helm upgrade --install memcached-release bitnami/memcached \
--set resources.requests.memory="100Mi",resources.requests.cpu="100m"

echo "deploying mysql"

helm upgrade --install mysql-release bitnami/mysql \
--set auth.rootPassword=password123