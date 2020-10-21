FROM debian:10-slim
RUN useradd -u 20001 postgres_exporter

# Install certs and create home directory needed by the AWS SDK
RUN apt update && apt install ca-certificates -y \
  && mkdir /home/postgres_exporter \
  && chown postgres_exporter:postgres_exporter /home/postgres_exporter \
  && rm -rf /var/lib/{apt,dpkg,cache,log}/

USER postgres_exporter

ARG binary

COPY $binary /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
