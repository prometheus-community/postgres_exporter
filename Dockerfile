FROM alpine
RUN set -x \
    && apk --update upgrade \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*
COPY postgres_exporter  /postgres_exporter
USER nobody:nobody

EXPOSE 9187
ENTRYPOINT  ["/postgres_exporter"]
