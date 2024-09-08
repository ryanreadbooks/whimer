FROM golang:1.22 AS builder
ARG PROJECT

WORKDIR /app

ENV CGO_ENABLED=0 GO111MODULE=on GOPROXY='https://goproxy.cn,direct'
COPY . .
RUN cd ${PROJECT} && go mod tidy && go build -ldflags='-s -w' -o ${PROJECT} cmd/main.go

FROM ubuntu:22.04
ARG PROJECT
WORKDIR /app

COPY --from=builder /app/${PROJECT}/etc /app
COPY --from=builder /app/${PROJECT}/${PROJECT} /app/server

ENTRYPOINT [ "/app/server" ]
