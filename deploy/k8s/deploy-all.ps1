# Deploy infrastructure
echo "Deploying infrastructure..."
kubectl apply -f infrastructure/consul/
kubectl apply -f infrastructure/redis/
kubectl apply -f infrastructure/mysql/
kubectl apply -f infrastructure/jaeger/
kubectl apply -f infrastructure/elk/

# Deploy microservices
echo "Deploying microservices..."
kubectl apply -f microservices/interface/
kubectl apply -f microservices/user/
kubectl apply -f microservices/task/
kubectl apply -f microservices/job/

# Deploy monitoring components
echo "Deploying monitoring components..."
kubectl apply -f monitoring/

echo "Deployment completed!"
