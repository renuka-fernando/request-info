FROM scratch
ARG TARGETARCH

COPY request-info-$TARGETARCH /bin/request-info
ENTRYPOINT ["/bin/request-info"]
EXPOSE 8080
