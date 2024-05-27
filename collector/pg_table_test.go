package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGTableSizeCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	rows := sqlmock.NewRows([]string{"datname", "relname", "schemaname", "total_relation_size", "relation_size", "indexes_size"}).
		AddRow("test", "testrel", "testschema", 69, 42, 27).
		AddRow("test2", "testrel2", "testschema2", 14, 10, 2)

	mock.ExpectQuery(sanitizeQuery(pgTableSizeQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGTableSizeCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGTableSizeCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "test", "relname": "testrel", "schemaname": "testschema"}, value: 69, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"datname": "test", "relname": "testrel", "schemaname": "testschema"}, value: 42, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"datname": "test", "relname": "testrel", "schemaname": "testschema"}, value: 27, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"datname": "test2", "relname": "testrel2", "schemaname": "testschema2"}, value: 14, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"datname": "test2", "relname": "testrel2", "schemaname": "testschema2"}, value: 10, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"datname": "test2", "relname": "testrel2", "schemaname": "testschema2"}, value: 2, metricType: dto.MetricType_COUNTER},
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
