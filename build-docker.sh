#!/bin/bash

env GOOS=linux GOARCH=386 go build -o request-info .
docker build . -t cakebakery/request-info:v1