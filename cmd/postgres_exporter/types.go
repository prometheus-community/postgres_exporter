//go:generate sh -c "test mocks/mocks.go -nt $GOFILE && exit 0; mockgen -destination mocks/mocks.go -package mocks . ServerAPI"

//   go:generate sh -c "test mocks/aws_mocks.go -nt $GOFILE && exit 0; mockgen -destination mocks/aws_mocks.go -package mocks . CloudWatchAPI,RDSAPI"
//   go:generate sh -c "test mocks/mocks.go -nt $GOFILE && exit 0; mockgen -destination mocks/mocks.go -package mocks . ServerAPI"

package postgres_exporter

import (
	"database/sql"

	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
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
	Close() error
	Scrape(ch chan<- prometheus.Metric) error
	Query(query string) (*sql.Rows, error)
}
