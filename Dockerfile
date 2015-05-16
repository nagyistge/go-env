FROM golang:1.4.2-cross
MAINTAINER peter.edge@gmail.com

RUN mkdir -p /go/src/github.com/peter-edge/go-env
ADD . /go/src/github.com/peter-edge/go-env/
WORKDIR /go/src/github.com/peter-edge/go-env
