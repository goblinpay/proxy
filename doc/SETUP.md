# Production setup

## setting up the db pass secret

```
kubectl create secret generic credentials --from-literal=db_pass="password123"
```

## setting up secret for CloudSQL proxy

```
kubectl create secret generic cloudsql-instance-credentials --from-file=credentials.json=$HOME/Downloads/goblin-services-13761ab4e518.json 
```
