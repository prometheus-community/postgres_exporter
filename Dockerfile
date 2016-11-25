FROM scratch

COPY postgres_exporter /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
