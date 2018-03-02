# sudo docker build -t negbie/heplify-server:latest .

FROM golang:latest as builder
RUN go get -d -v -u github.com/negbie/heplify-server/
WORKDIR /go/src/github.com/negbie/heplify-server/cmd/heplify-server/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -installsuffix cgo -o heplify-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/negbie/heplify-server/cmd/heplify-server/heplify-server .
EXPOSE 9060
CMD ["./heplify-server", "-h"]