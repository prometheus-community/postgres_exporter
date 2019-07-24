FROM debian:7.11-slim
RUN useradd -u 20001 postgres_exporter

FROM scratch

COPY --from=0 /etc/passwd /etc/passwd
USER postgres_exporter

ARG binary

COPY $binary /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
