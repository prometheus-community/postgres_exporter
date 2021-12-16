package postgres_exporter_test

import (
	"reflect"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/everestsystems/postgres_exporter/cmd/postgres_exporter"
)

var _ = Describe("PgSetting", func() {
	Context("QueryNamespaceMappings", func() {
		var (
			ctrl *gomock.Controller
			s    SettingsMetrics

			serverLabels prometheus.Labels

			metricDesc []string
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			serverLabels = prometheus.Labels{
				ServerLabelName: "hostname:5432",
			}

			metricDesc = []string{
				`Desc{fqName: "pg_settings_archive_timeout_seconds", help: "aaa [Units converted to seconds.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_autovacuum_vacuum_cost_delay_seconds", help: "aaa [Units converted to seconds.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_autovacuum_work_mem_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_effective_cache_size_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_log_rotation_age_seconds", help: "aaa [Units converted to seconds.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_pg_stat_statements_save", help: "aaa", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_seq_page_cost", help: "aaa", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_seconds", help: "aaa [Units converted to seconds.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_seconds", help: "aaa [Units converted to seconds.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
				`Desc{fqName: "pg_settings_dummy_bytes", help: "aaa [Units converted to bytes.]", constLabels: {server="hostname:5432"}, variableLabels: []}`,
			}

		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should pass if QuerySettings works", func() {
			db, mock, _ := sqlmock.New()

			defer db.Close()

			rows := sqlmock.NewRows([]string{"name", "setting", "coalesce", "short_desc", "vartype"}).
				AddRow("archive_timeout", "300", "s", "aaa", "integer").
				AddRow("autovacuum_vacuum_cost_delay", "5", "ms", "aaa", "integer").
				AddRow("autovacuum_work_mem", "1", "kB", "aaa", "integer").
				AddRow("effective_cache_size", "271158", "8kB", "aaa", "integer").
				AddRow("log_rotation_age", "271158", "min", "aaa", "integer").
				AddRow("pg_stat_statements.save", "on", "", "aaa", "bool").
				AddRow("seq_page_cost", "1", "", "aaa", "real").
				AddRow("dummy", "1", "h", "aaa", "integer").
				AddRow("dummy", "1", "d", "aaa", "integer").
				AddRow("dummy", "1", "MB", "aaa", "integer").
				AddRow("dummy", "1", "GB", "aaa", "integer").
				AddRow("dummy", "1", "TB", "aaa", "integer").
				AddRow("dummy", "1", "8kB", "aaa", "integer").
				AddRow("dummy", "1", "16kB", "aaa", "integer").
				AddRow("dummy", "1", "32kB", "aaa", "integer").
				AddRow("dummy", "1", "16MB", "aaa", "integer").
				AddRow("dummy", "1", "32MB", "aaa", "integer").
				AddRow("dummy", "1", "64MB", "aaa", "integer")

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real')`)).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				err := s.QuerySettings(ch, db, serverLabels)
				if err != nil {
					return
				}
				close(ch)
			}()

			for {
				res, ok := <-ch
				if ok == false {
					break
				}
				ret := prometheus.Metric(res)
				list = append(list, ret.Desc().String())
			}

			Expect(reflect.DeepEqual(list, metricDesc)).To(BeTrue())
		})

		It("should fail if Query fail", func() {
			db, mock, _ := sqlmock.New()

			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real')`)).WillReturnError(errDummy)

			ch := make(chan prometheus.Metric)

			err := s.QuerySettings(ch, db, serverLabels)

			Expect(err).To(MatchError(err))
		})

	})
})
