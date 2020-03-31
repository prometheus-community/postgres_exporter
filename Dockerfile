FROM debian:10-slim
RUN useradd -u 20001 postgres_exporter

USER postgres_exporter

ARG binary

COPY $binary /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
