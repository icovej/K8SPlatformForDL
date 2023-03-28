WORKPATH = $(shell pwd)/PlatformBackEnd

.PHONY: clean

build:
	CGO_ENABLED=0 GOOS=linux go run ${WORKPATH}/main.go srcfilepath=${WORKPATH}/dockerfile -log_dir=${WORKPATH}/logs -alsologtostderr > /dev/null 2>&1 &

