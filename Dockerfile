FROM golang:alpine as builder

RUN apk update && apk add --no-cache git
RUN mkdir -p /go/src/cli
COPY . /go/src/cli
WORKDIR /go/src/cli
RUN go get && go mod tidy
RUN go build -o bima .
RUN mv /go/src/cli/bima /usr/local/bin/bima
RUN chmod a+x /usr/local/bin/bima
