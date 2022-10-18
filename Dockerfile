FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /go
COPY build_artifact_bin testapi
RUN chmod 755 testapi
COPY config.yml config.yml

EXPOSE 8080
ENTRYPOINT ["/go/testapi"]
