# should be run from root directory of heplify-server
# sudo docker build --no-cache -t sipcapture/heplify-server -f docker/Dockerfile . 
FROM golang:alpine as builder

RUN apk update && apk add --no-cache git build-base 
RUN git clone https://luajit.org/git/luajit-2.0.git \
 && cd luajit-2.0 \
 && git checkout v2.1 \
 && make CCOPT="-static -fPIC" BUILDMODE="static" && make install
COPY . /heplify-server
WORKDIR /heplify-server/cmd/heplify-server
RUN CGO_ENABLED=1 GOOS=linux go build -a --ldflags '-linkmode external -extldflags "-static -s -w"' -o heplify-server .

FROM alpine:latest

RUN apk update
RUN apk --no-cache add ca-certificates bash
RUN apk upgrade

WORKDIR /root/
COPY --from=builder /heplify-server/cmd/heplify-server/heplify-server .
COPY --from=builder /heplify-server/scripts ./scripts
CMD ["./heplify-server"]
