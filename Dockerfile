FROM scratch

ARG binary

COPY $binary /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
