cluster:
	kubectl apply -Rf kubernetes/

build:
	docker build -t goblin/proxy .

push:
	docker tag goblin/proxy gcr.io/goblin-proxy-238316/proxy:0.0
	docker push gcr.io/goblin-proxy-238316/proxy:0.0

deploy:
	kubectl apply -f deployment.yaml
