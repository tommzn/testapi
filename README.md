# Test API
An API which logs all incoming requests.
## Kubernetes 
```
kubectl apply -f k8s/deployment.yml
```
## Docker
```
docker run -d -p 8080:31475 tommzn/testapi:latest
```

