FROM alpine:3.8

CMD docserver

# we connect to S3 over HTTPS
RUN apk add --no-cache ca-certificates \
	&& mkdir -p /docserver/_packages

ADD src/docserver /bin/docserver
