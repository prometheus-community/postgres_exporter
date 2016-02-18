FROM golang:1.4.2
ADD . /go/src/app
RUN cd /go/src/app && make

EXPOSE 9113
ENTRYPOINT [ "/go/src/app/postgres_exporter" ]