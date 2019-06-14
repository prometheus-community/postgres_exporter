FROM golang AS build
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor GOPROXY=off go build

FROM scratch
COPY --from=build /src/postgres_exporter /postgres_exporter
EXPOSE 9187
ENTRYPOINT ["/postgres_exporter"]
