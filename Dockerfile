FROM alpine:3.5

# we connect to S3 over HTTPS
RUN apk add --no-cache ca-certificates \
	&& mkdir -p /docserver/_packages

ADD src/docserver /docserver

CMD /docserver
