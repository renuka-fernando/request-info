#!/bin/bash

# install docker buildx plugin

env GOOS=linux GOARCH=amd64 go build -o request-info-amd64 .
env GOOS=linux GOARCH=arm64 go build -o request-info-arm64 .

docker buildx build --push --platform linux/arm64/v8,linux/amd64 -t cakebakery/request-info:v1 .
#docker build --build-arg TARGETARCH=arm64 -t cakebakery/request-info:v1 .

