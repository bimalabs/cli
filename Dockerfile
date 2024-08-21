FROM ubuntu:latest as builder

RUN apt update && apt install -y git golang gcc libc-dev
RUN mkdir -p /go/src/cli
WORKDIR /go/src/cli
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o bima .
RUN mv /go/src/cli/bima /usr/local/bin/bima
RUN chmod a+x /usr/local/bin/bima
RUN bima version

FROM ubuntu:latest

COPY --from=builder /usr/local/bin/bima /usr/local/bin/bima
RUN chmod a+x /usr/local/bin/bima
RUN bima version
