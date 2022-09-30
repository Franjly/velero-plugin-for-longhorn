FROM golang:1.19.1-alpine AS builder
ENV GOPROXY=https://proxy.golang.org
WORKDIR /go/src/github.com/ecatlabs/velero-plugin
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/velero-plugin .

FROM ubuntu:jammy
RUN mkdir /plugins
COPY --from=builder /go/bin/velero-plugin /plugins/
USER nobody:nogroup
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
