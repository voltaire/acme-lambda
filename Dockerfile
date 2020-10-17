FROM golang:alpine AS builder
ADD . /src/
WORKDIR /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /controller github.com/voltaire/map-cert/controller

FROM alpine:latest
COPY --from=builder /controller /etc/periodic/monthly/controller
RUN apk --no-cache add tzdata openntpd && \
    cp /usr/share/zoneinfo/UTC /etc/localtime && \
    echo "UTC" > /etc/timezone
ADD start /start
CMD ["/bin/sh", "/start"]