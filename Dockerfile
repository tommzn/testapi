FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /go

COPY build_artifact_bin testapi
COPY config.yml config.yml

RUN chmod 755 /go/testapi
ENTRYPOINT ["/go/testapi"]
