ARG ARCH="amd64"
ARG OS="linux"
FROM golang:1.18-alpine AS build-env

ARG APPNAME
ENV GO111MODULE=auto
ENV SRCPATH $GOPATH/src/github.com/form3tech-oss/$APPNAME

COPY ./ $SRCPATH

RUN go install github.com/form3tech-oss/$APPNAME/cmd/$APPNAME


FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

COPY --from=build-env /go/bin/$APPNAME /bin/

EXPOSE     9187
USER       nobody
ENTRYPOINT [ "/bin/postgres_exporter" ]
