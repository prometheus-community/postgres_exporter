FROM --platform=arm64 golang:1.17.5-alpine3.15 as stage
ENV ARCH="arm64"
ENV OS="linux"

RUN apk add git make curl && \ 
        git clone https://github.com/everestsystems/postgres_exporter.git && \
        pwd && \
        cd /go/postgres_exporter && \
        make build

FROM --platform=arm64 golang:1.17.5-alpine3.15
ENV ARCH="arm64"
ENV OS="linux"

COPY --from=stage  /go/postgres_exporter/.build/postgres_exporter . 

EXPOSE     9187
USER       nobody

ENTRYPOINT [ "/go/postgres_exporter" ]
