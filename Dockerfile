FROM golang:1.13.2 AS builder

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG GO111MODULE=on

COPY . /src
RUN cd /src && go build -mod=vendor -o /tmp/cloud-run-grpc-concurrency-test ./server

FROM gcr.io/distroless/base

COPY --from=builder /tmp/cloud-run-grpc-concurrency-test /usr/bin/

ENTRYPOINT ["/usr/bin/cloud-run-grpc-concurrency-test" ]
