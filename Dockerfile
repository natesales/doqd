FROM alpine
COPY doqd /usr/bin/doqd
ENTRYPOINT ["/usr/bin/doqd"]