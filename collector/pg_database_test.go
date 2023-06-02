package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGDatabaseCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	mock.ExpectQuery(sanitizeQuery(pgDatabaseQuery)).WillReturnRows(sqlmock.NewRows([]string{"datname"}).
		AddRow("postgres"))

	mock.ExpectQuery(sanitizeQuery(pgDatabaseSizeQuery)).WithArgs("postgres").WillReturnRows(sqlmock.NewRows([]string{"pg_database_size"}).
		AddRow(1024))

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGDatabaseCollector{}
		if err := c.Update(context.Background(), db, ch); err != nil {
			t.Errorf("Error calling PGDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres"}, value: 1024, metricType: dto.MetricType_GAUGE},
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
