FROM ubuntu:latest

COPY wavefront-collector /
COPY ca-certificates.crt /etc/ssl/certs/

#   nobody:nobody
USER 65534:65534
ENTRYPOINT ["/wavefront-collector"]
