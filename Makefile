SHELL=/bin/bash
dev: build-local deploy-local
production: build deploy

build:
	rm -f ./bin/*
	go get
	CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/router -installsuffix cgo .
	docker build -t agxp/router .

build-local:
	rm -f ./bin/*
	go get
	CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/router -installsuffix cgo .
	@eval $$(minikube docker-env) ;\
	docker build -t router .


run:
	docker run --net="host" \
		-p 50051 \
		-e MICRO_SERVER_ADDRESS=:50051 \
		-e MICRO_REGISTRY=mdns \
		-e MINIO_URL=minio-0 \
		-e MINIO_ACCESS_KEY=minio \
		-e MINIO_SECRET_KEY=minio123 \
		router

deploy:
	docker push agxp/router
	sed "s/{{ UPDATED_AT }}/$(shell date)/g" ./deployments/deployment.tmpl > ./deployments/deployment.yaml
	kubectl apply -f ./deployments/deployment.yaml

deploy-local:
	sed "s,{{ MINIO_EXTERNAL_URL }},$(shell minikube service minio --url),g" ./deployments/deployment.tmpl > ./deployments/deployment.tmpl1
	sed "s/{{ UPDATED_AT }}/$(shell date)/g" ./deployments/deployment.tmpl1 > ./deployments/deployment.yaml
	kubectl apply -f ./deployments/deployment.yaml
	kubectl apply -f ./deployments/service.yaml
