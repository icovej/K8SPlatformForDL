FROM golang:1.17.3 as builder
ARG BUILD_SHA
ARG BUILD_VERSION
WORKDIR /root/.go/src/metagpu
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY gen gen
RUN go mod tidy
RUN go build \
    -ldflags="-extldflags=-Wl,-z,lazy -s -w -X 'main.Build=${BUILD_SHA}' -X 'main.Version=${BUILD_VERSION}'" \
    -o mgdp cmd/mgdp/main.go
RUN go build \
     -ldflags="-X 'main.Build=${BUILD_SHA}' -X 'main.Version=${BUILD_VERSION}'" \
     -o mgctl cmd/mgctl/*.go
RUN go build \
     -ldflags="-X 'main.Build=${BUILD_SHA}' -X 'main.Version=${BUILD_VERSION}'" \
     -o mgex cmd/mgex/*.go

FROM nvidia/cuda:11.6.0-base-ubuntu20.04

ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=utility

LABEL io.k8s.display-name="cnvrg.io Meta GPU Device Plugin"
LABEL name="cnvrg.io MetaGPU Device Plugin"
LABEL vendor="cnvrg.io"
LABEL version="N/A"
LABEL release="N/A"
LABEL summary="cnvrg.io MetaGPU device plugin for Kubernetes"
LABEL description="See summary"
COPY --from=builder /root/.go/src/metagpu/mgdp /usr/bin/mgdp
COPY --from=builder /root/.go/src/metagpu/mgctl /usr/bin/mgctl
COPY --from=builder /root/.go/src/metagpu/mgex /usr/bin/mgex
RUN cp /usr/bin/mgctl /tmp