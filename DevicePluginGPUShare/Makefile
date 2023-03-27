IMAGE_VERSION = latest
REGISTRY = icovej
IMAGE = ${REGISTRY}/new-device-plugin:${IMAGE_VERSION}

.PHONY: build deploy

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/dn main.go

image:
	docker build -t ${IMAGE} .

push:
	docker push ${IMAGE}