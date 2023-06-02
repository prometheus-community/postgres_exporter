package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgReplicationSlotCollectorActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	columns := []string{"slot_name", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow("test_slot", 5, 3, true)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), db, ch); err != nil {
			t.Errorf("Error calling PGPostmasterCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot"}, value: 5, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot"}, value: 3, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot"}, value: 1, metricType: dto.MetricType_GAUGE},
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

func TestPgReplicationSlotCollectorInActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	columns := []string{"slot_name", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow("test_slot", 6, 12, false)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), db, ch); err != nil {
			t.Errorf("Error calling PGPostmasterCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot"}, value: 6, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot"}, value: 0, metricType: dto.MetricType_GAUGE},
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
