# Docker Images

Official images are published to three registries, kept in sync on every release:

- Docker Hub: [`prometheuscommunity/postgres-exporter`](https://hub.docker.com/r/prometheuscommunity/postgres-exporter)
- Quay.io: [`quay.io/prometheuscommunity/postgres-exporter`](https://quay.io/repository/prometheuscommunity/postgres-exporter)
- GHCR: [`ghcr.io/prometheus-community/postgres-exporter`](https://github.com/prometheus-community/postgres_exporter/pkgs/container/postgres-exporter)

All three are the same image; there's no functional difference between them — use whichever registry your infrastructure prefers. Note that the GHCR organization is `prometheus-community` (hyphenated, matching the GitHub org), while Docker Hub and Quay use `prometheuscommunity`.

## Tags

- `latest` — the most recent release.
- `vX.Y.Z` — a specific release, e.g. `v0.18.1`. Pin to a specific tag for production deployments so upgrades are intentional.
- `master` — the latest build from the `master` branch, which may be unstable. Use this only for testing or development.

Available tags are listed on the [Docker Hub tags page](https://hub.docker.com/r/prometheuscommunity/postgres-exporter/tags).

## Image details

- Based on a minimal `busybox` image — there is no shell to `exec` into.
- The exporter binary is at `/bin/postgres_exporter` and is also the container `ENTRYPOINT`, so flags can be passed directly as container arguments.
- Exposes port `9187`.
- The process runs as the `nobody` user, uid/gid `65534`. If you mount config or secret files into the container (e.g. for `DATA_SOURCE_PASS_FILE` or `--config.file`), make sure they're readable by that uid/gid.

## Running

```bash
docker run \
  --net=host \
  -e DATA_SOURCE_URI="localhost:5432/postgres?sslmode=disable" \
  -e DATA_SOURCE_USER=postgres \
  -e DATA_SOURCE_PASS=password \
  quay.io/prometheuscommunity/postgres-exporter
```

To mount a secrets file instead of passing a password in an environment variable:

```bash
docker run \
  -p 9187:9187 \
  -e DATA_SOURCE_URI="my-postgres-host:5432/postgres?sslmode=disable" \
  -e DATA_SOURCE_USER=postgres \
  -e DATA_SOURCE_PASS_FILE=/run/secrets/pg_password \
  -v /path/to/pg_password:/run/secrets/pg_password:ro \
  quay.io/prometheuscommunity/postgres-exporter
```

See [Secrets](secrets.md) for all the supported ways to supply credentials.

## Building the image yourself

```bash
make promu
promu crossbuild -p linux/amd64 -p linux/armv7 -p linux/arm64 -p linux/ppc64le
make docker
```

This produces a local image tagged `prometheuscommunity/postgres_exporter:${branch}`.
