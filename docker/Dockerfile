# sudo docker build -t sipcapture/heplify-server:latest .

FROM golang:alpine as builder

RUN apk update && apk add --no-cache git
RUN go get -u -d -v github.com/negbie/heplify-server/...
WORKDIR /go/src/github.com/negbie/heplify-server/cmd/heplify-server/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -installsuffix cgo -o heplify-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/negbie/heplify-server/cmd/heplify-server/heplify-server .
