module github.com/prometheus-community/postgres_exporter

go 1.14

require (
	github.com/aws/aws-sdk-go-v2 v1.11.2
	github.com/aws/aws-sdk-go-v2/config v1.11.0
	github.com/aws/aws-sdk-go-v2/credentials v1.6.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/rds v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.11.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-kit/kit v0.11.0
	github.com/lib/pq v1.10.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.29.0
	github.com/prometheus/exporter-toolkit v0.6.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 v2.4.0
)
