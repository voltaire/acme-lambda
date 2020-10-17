FROM golang:alpine
WORKDIR /src/
ADD . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /controller github.com/voltaire/map-cert/controller
ENTRYPOINT [ "/controller" ]