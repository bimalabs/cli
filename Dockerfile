FROM golang:alpine as builder

RUN apk update && apk add --no-cache git
RUN mkdir -p /go/src/cli
WORKDIR /go/src/cli
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o bima .
RUN mv /go/src/cli/bima /usr/local/bin/bima
RUN chmod a+x /usr/local/bin/bima

FROM golang:alpine

COPY --from=builder /usr/local/bin/bima /usr/local/bin/bima
