FROM alpine:3.8

CMD ["docserver"]

# we connect to S3 over HTTPS
RUN apk add --no-cache ca-certificates \
	&& mkdir -p /docserver/_packages

ADD rel/docserver_linux-amd64 /usr/local/bin/docserver
