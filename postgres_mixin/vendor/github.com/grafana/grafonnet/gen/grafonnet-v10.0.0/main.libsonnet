local grafonnet = import 'github.com/grafana/grafonnet/grafonnet-base/main.libsonnet';
local schemas = import './schemas.libsonnet';
local version = importstr './grafana-version';
grafonnet.new(schemas, version)
