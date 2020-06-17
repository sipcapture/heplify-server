# sudo docker build -t sipcapture/heplify-server:latest .

FROM golang:alpine as builder

RUN apk update && apk add --no-cache git build-base

RUN git clone https://luajit.org/git/luajit-2.0.git \
 && cd luajit-2.0 \
 && make CCOPT="-static -fPIC" BUILDMODE="static" && make install

RUN go get -u -d -v github.com/sipcapture/heplify-server/...
WORKDIR /go/src/github.com/sipcapture/heplify-server/cmd/heplify-server/
RUN CGO_ENABLED=1 GOOS=linux go build -a --ldflags '-linkmode external -extldflags "-static -s -w"' -o heplify-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/sipcapture/heplify-server/cmd/heplify-server/heplify-server .
COPY --from=builder /go/src/github.com/sipcapture/heplify-server/scripts ./scripts
