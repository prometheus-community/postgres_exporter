package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatIOUserTablesCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	columns := []string{
		"datname",
		"schemaname",
		"relname",
		"heap_blks_read",
		"heap_blks_hit",
		"idx_blks_read",
		"idx_blks_hit",
		"toast_blks_read",
		"toast_blks_hit",
		"tidx_blks_read",
		"tidx_blks_hit",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres",
			"public",
			"a_table",
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8)
	mock.ExpectQuery(sanitizeQuery(statioUserTablesQuery)).WillReturnRows(rows)
	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatIOUserTablesCollector{}

		if err := c.Update(context.Background(), db, ch); err != nil {
			t.Errorf("Error calling PGStatIOUserTablesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 1},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 2},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 3},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 4},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 6},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 7},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 8},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
