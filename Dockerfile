FROM --platform=$BUILDPLATFORM golang:1.24 AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /req-info

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/request-info .

FROM alpine:3 AS final

RUN apk update && apk upgrade --no-cache
RUN apk add --no-cache tzdata

USER 10014

COPY --from=build /go/bin/request-info /bin/request-info
EXPOSE 8080
ENTRYPOINT ["/bin/request-info"]
