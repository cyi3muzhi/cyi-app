# Deploy infrastructure
echo "Deploying infrastructure..."
kubectl apply -f infrastructure/consul/
kubectl apply -f infrastructure/redis/
kubectl apply -f infrastructure/mysql/
kubectl apply -f infrastructure/jaeger/

# Deploy ELK stack
# echo "Deploying ELK stack..."
kubectl apply -f infrastructure/elk/logstash/
kubectl apply -f infrastructure/elk/elasticsearch/
kubectl apply -f infrastructure/elk/kibana/
kubectl apply -f infrastructure/elk/filebeat/

# Deploy microservices
echo "Deploying microservices..."
kubectl apply -f microservices/blades/
kubectl apply -f microservices/interface/
kubectl apply -f microservices/user/
kubectl apply -f microservices/task/
kubectl apply -f microservices/job/

# Deploy monitoring components
echo "Deploying monitoring components..."
kubectl apply -f monitoring/prometheus/
kubectl apply -f monitoring/grafana/

echo "Deployment completed!"
