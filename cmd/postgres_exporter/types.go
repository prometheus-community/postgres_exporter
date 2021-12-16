//go:generate sh -c "test mocks/mocks.go -nt $GOFILE && exit 0; mockgen -destination mocks/mocks.go -package mocks . ServerAPI,RDSMetricsAPI,NamespaceMetricsAPI,SettingsMetricsAPI"
//go:generate sh -c "test mocks/aws_mocks.go -nt $GOFILE && exit 0; mockgen -destination mocks/aws_mocks.go -package mocks . CloudWatchAPI,RDSAPI"

package postgres_exporter

import (
	"database/sql"

	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// HTTPHandler wrapper the native http.Handler interface for using within gomock.
type CloudWatchAPI interface {
	cloudwatchiface.CloudWatchAPI
}

// RDSAPI is used for mocking the SQS interface.
type RDSAPI interface {
	rdsiface.RDSAPI
}

// ServerAPI interface for database operations
type ServerAPI interface {
	Open() error
	Close() error
	Scrape(ch chan<- prometheus.Metric, totalScrapes, rdsDatabaseConnections, rdsCurrentCapacity float64) error
	Query(query string) (*sql.Rows, error)
}

// RDSMetricsAPI interface for database metrics
type RDSMetricsAPI interface {
	RdsCurrentCapacity(clusterID string) (int64, error)
	RdsCurrentConnections(clusterID string) (int64, error)
}

// NamespaceMetricsAPI interface for collect server metrics
type NamespaceMetricsAPI interface {
	QueryNamespaceMappings(ch chan<- prometheus.Metric, db *sql.DB, serverLabels prometheus.Labels, queryList map[string]string,
		metricMaps map[string]IntermediateMetricMap, semanticVersion semver.Version, versionString string) map[string]error
	SetInternalMetrics(ch chan<- prometheus.Metric, duration, totalScrapes, rdsDatabaseConnections, rdsCurrentCapacity float64)
}

type SettingsMetricsAPI interface {
	QuerySettings(ch chan<- prometheus.Metric, db *sql.DB, serverLabels prometheus.Labels) error
}
