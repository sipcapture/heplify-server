FROM golang:alpine as builder

LABEL maintainer="zhouxq@ti-net.com.cn"

RUN set -ex \
  && apk add --update  bash\
  && rm -rf /var/cache/apk/*

ENV BUILD_ROOT=/go/src/github.com/sipcapture/heplify-server/
COPY . ${BUILD_ROOT} 

RUN cd ${BUILD_ROOT} \
  && cp vendor/* /go/src/ -rf \
  && cd ${BUILD_ROOT}/cmd/heplify-server \
  && CGO_ENABLED=0 GOOS=linux go build -v -a -ldflags '-s -w' -installsuffix cgo -o  /go/bin/heplify-server \
  && rm -rf /go/src/

FROM alpine:latest
LABEL maintainer="zhouxq@ti-net.com.cn"

RUN apk --no-cache add ca-certificates m4 bash

WORKDIR /root/

COPY --from=builder /go/bin/heplify-server .
COPY docker/files/entrypoint.sh /entrypoint.sh
COPY docker/files/heplify-server.toml.m4 /root/heplify-server.toml.m4

RUN chmod 744 /root/heplify-server.toml.m4 \
  && chmod a+rx /entrypoint.sh

EXPOSE 9060

CMD ["/entrypoint.sh"]