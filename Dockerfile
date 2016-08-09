FROM alpine:3.4

RUN apk add --update tini

ADD postgres_exporter /

EXPOSE 9113

ENTRYPOINT ["/sbin/tini", "--", "/postgres_exporter"]