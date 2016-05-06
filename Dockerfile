FROM golang:1.6-alpine

RUN apk --update add make

COPY . /go/src/github.com/wrouesnel/postgres_exporter

WORKDIR /go/src/github.com/wrouesnel/postgres_exporter

RUN make postgres_exporter

ENTRYPOINT [ "./postgres_exporter" ]

EXPOSE 9113
