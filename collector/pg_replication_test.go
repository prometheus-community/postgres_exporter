package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgReplicationCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationCollector{}

		if err := c.Update(context.Background(), db, ch); err != nil {
			t.Errorf("Error calling PGPostmasterCollector.Update: %s", err)
		}
	}()

	columns := []string{"lag", "is_replica"}
	rows := sqlmock.NewRows(columns).
		AddRow(1000, 1)
	mock.ExpectQuery(sanitizeQuery(pgReplicationQuery)).WillReturnRows(rows)

	expected := []MetricResult{
		{labels: labelMap{}, value: 1000, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 1, metricType: dto.MetricType_GAUGE},
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
