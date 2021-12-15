package postgres_exporter_test

import (
	"reflect"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter"
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
			s = SettingsMetrics{}

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
				AddRow("seq_page_cost", "1", "", "aaa", "real")

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real')`)).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				s.QuerySettings(ch, db, serverLabels)
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

	})
})
