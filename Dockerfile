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

COPY --from=stage  /go/postgres_exporter/postgres_exporter . 

ENV PG_IAM_ROLE_ARN=arn:aws:iam::680189258452:role/tenant-cff0019d-41c0-46fd-8c8f-cc6ee9786162
ENV PG_TENANT_ID=cff0019d-41c0-46fd-8c8f-cc6ee9786162
ENV PG_CLUSTER_ID=tenant-cff0019d-41c0-46fd-8c8f-cc6ee9786162

EXPOSE     9187
#USER       nobody

ENTRYPOINT [ "/go/postgres_exporter" ]
