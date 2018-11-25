FROM golang:alpine as builder

WORKDIR /heplify-server
COPY . /heplify-server

RUN apk update && apk add --no-cache git
RUN env GO111MODULE=off go get -u github.com/gobuffalo/packr/... && \
    /go/bin/packr -i /heplify-server/database

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -installsuffix cgo -o heplify-server ./cmd/heplify-server

FROM alpine:latest
RUN apk update && apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /heplify-server/heplify-server .
CMD ["./heplify-server"]