FROM golang:1.19-alpine as build

WORKDIR /home/gpu-server/set_k8s/biyesheji/DevicePluginGPUShare

COPY . .

# RUN go build -o /home/gpu-server/set_k8s/biyesheji/DevicePluginGPUShare/gpushare-sche-extender cmd/*.go

FROM alpine

COPY --from=build /home/gpu-server/set_k8s/biyesheji/DevicePluginGPUShare/gpushare-sche-extender /usr/bin/gpushare-sche-extender

CMD ["gpushare-sche-extender"]
