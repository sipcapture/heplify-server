# The release binaries are static. This image only uploads them.
# debian:bullseye is archived, so its security pool returns 404.
FROM debian:trixie-slim

RUN apt-get update && \
    apt-get install --yes --no-install-recommends bash ca-certificates curl jq && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY upload-script /usr/bin/upload-script
RUN chmod +x /usr/bin/upload-script

ENTRYPOINT ["upload-script"]
