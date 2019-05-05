CACHE_DIR=.makecache
DIGEST_FILE=$(CACHE_DIR)/DIGEST

# unorganized or tests
create-dir:
	mkdir -p $(CACHE_DIR)

run:
	docker run -it -p 8181:8080 goblin/proxy

cluster: 
	kubectl apply -Rf kubernetes/



build:
	docker build -t goblin/proxy .

tag: 
	docker tag goblin/proxy gcr.io/goblin-proxy-238316/proxy:0.0

push:
	docker push gcr.io/goblin-proxy-238316/proxy:0.0 | tee /dev/tty | awk '/digest: / {print $$3 > "$(DIGEST_FILE)"}' 

deploy:
	$(eval DIGEST=$(shell cat $(DIGEST_FILE)))
	DIGEST=$(DIGEST) envsubst <kubernetes/deployment.yaml | kubectl apply -f -

full: build tag push deploy



.PHONY: cluster build tag push deploy full



## remote build & push

# gcloud config set project goblin-proxy-238316
# gcloud builds submit --tag gcr.io/goblin-proxy-238316/proxy
