FROM --platform=$BUILDPLATFORM pscale.dev/wolfi-prod/go:1.23 AS build
ARG TARGETOS
ARG TARGETARCH
RUN apk --no-cache add curl inotify-tools
COPY . /postgres_exporter
RUN rm -f /postgres_exporter/postgres_exporter
RUN CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" make -C /postgres_exporter build

FROM pscale.dev/wolfi-prod/base:latest
COPY --from=build /postgres_exporter/postgres_exporter /bin/postgres_exporter
COPY --from=build /bin/inotifywait /bin/inotifywait
COPY --from=build /lib/libinotifytools.so.0 /lib/libinotifytools.so.0
EXPOSE 9187
USER nobody
WORKDIR /
ENTRYPOINT ["/bin/postgres_exporter"]
