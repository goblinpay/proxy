# Goblin proxy

## running locally

```
go run proxy
```

## building container

```
docker build -t goblin/proxy .
```

## running container

```
docker run -it -p 8181:8080 goblin/proxy
```

## tag & push image

```
docker tag goblin/proxy gcr.io/goblin-proxy-238316/proxy:0.0
docker push gcr.io/goblin-proxy-238316/proxy:0.0
```

## remote build & push

```
gcloud config set project goblin-proxy-238316
gcloud builds submit --tag gcr.io/goblin-proxy-238316/proxy
```

## deploy setup

```
gcloud config set compute/zone us-central1-a
gcloud container clusters create proxy-cluster
gcloud container clusters get-credentials
```

## Inspect

```
kubectl get all
```

### list running nodes

```
kubectl get pods -l run=proxy-deployment -o wide
```

### logs

```
kubectl logs deployment/proxy-deployment proxy
```

## deploy

```
kubectl apply -f deployment.yaml
```

```
kubectl expose deployment/proxy-deployment --type LoadBalancer \
  --port 80 --target-port 8080
```

Or

```
kubectl apply -f service.yaml
```

## get load balancer IP

```
kubectl get service proxy-service --output yaml
```

## Deploying an update

### Deploy

```
kubectl apply -f deployment.yaml
```

### Check

```
kubectl get pods -l run=proxy -o wide
```

## Service/load balancer

### Create

```
kubectl apply -f service.yaml
```

### Check

```
kubectl get svc proxy-service
kubectl describe svc proxy-service
```

For basic registry commands, see: https://cloud.google.com/container-registry/docs/quickstart

- add config file
- add uid/uip logging, found
- improve performance/memory: use pointers appropriately
- add test, benchmarks, race test
- check out https://www.youtube.com/watch?v=1V7eJ0jN8-E for debug/optimization tips
