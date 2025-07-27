FROM golang:1.24 AS build
WORKDIR /req-info

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/request-info-amd64 .

FROM alpine:3 AS final

RUN apk update && apk upgrade --no-cache
RUN apk add --no-cache tzdata

USER 10014

COPY --from=build /go/bin/request-info-amd64 /bin/request-info
EXPOSE 8080
ENTRYPOINT ["/bin/request-info"]
