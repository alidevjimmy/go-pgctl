FROM alpine:3.14

RUN mkdir -p /config

COPY ../config/config.yaml /config/config.yaml
COPY ../build/go-pgctl /go-pgctl

CMD ["./go-pgctl"]