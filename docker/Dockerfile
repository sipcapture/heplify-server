# sudo docker build -t sipcapture/heplify-server:latest .

FROM golang:alpine as builder

RUN apk update && apk add --no-cache git build-base luajit-dev
RUN go install github.com/sipcapture/heplify-server/cmd/heplify-server@latest
#workaround for latest version
WORKDIR /go/pkg/mod/github.com/sipcapture/
RUN ln -s `ls -ltd hepl* | tail -1 | awk '{print $9}'` heplify-server@latest

FROM alpine:latest

RUN apk update
RUN apk --no-cache add ca-certificates
RUN apk upgrade

WORKDIR /root/
COPY --from=builder /go/bin/heplify-server .
COPY --from=builder /go/pkg/mod/github.com/sipcapture/heplify-server@latest/scripts ./scripts
