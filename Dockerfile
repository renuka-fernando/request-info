FROM alpine:3.18.0
ARG TARGETARCH

COPY request-info-$TARGETARCH /bin/request-info
ENTRYPOINT ["/bin/request-info"]
EXPOSE 8080
