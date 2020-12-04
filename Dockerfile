FROM alpine:3.12.1

COPY request-info /bin
ENTRYPOINT ["request-info"]
EXPOSE 8080