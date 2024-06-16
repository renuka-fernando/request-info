#!/bin/bash

IMAGE_NAME=renukafernando/request-info:latest

# install docker buildx plugin

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o request-info-amd64 .
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o request-info-arm64 .

#docker buildx build --push --platform linux/arm64/v8,linux/amd64 -t "$IMAGE_NAME" .
docker build --build-arg TARGETARCH=arm64 -t "$IMAGE_NAME" .
