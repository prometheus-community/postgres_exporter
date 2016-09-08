FROM scratch

COPY postgres_exporter /postgres_exporter

EXPOSE 9113

ENTRYPOINT [ "/postgres_exporter" ]
