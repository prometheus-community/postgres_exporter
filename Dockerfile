FROM golang:1.13-buster AS base
RUN useradd -u 20001 postgres_exporter

WORKDIR /go/src/app

COPY . .
RUN go run ./mage.go binary

FROM scratch

COPY --from=base /etc/passwd /etc/passwd
USER postgres_exporter

COPY --from=base /go/src/app/postgres_exporter /postgres_exporter

EXPOSE 9187

ENTRYPOINT [ "/postgres_exporter" ]
