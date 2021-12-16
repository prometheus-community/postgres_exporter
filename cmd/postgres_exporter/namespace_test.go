package postgres_exporter_test

import (
	"reflect"

	"github.com/blang/semver"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/gomega"
	. "github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter"
)

var _ = Describe("namespace", func() {
	Context("QueryNamespaceMappings", func() {
		var (
			ctrl *gomock.Controller

			nsMap              NamespaceMappings
			serverLabels       prometheus.Labels
			queries            map[string]string
			metricMaps         map[string]IntermediateMetricMap
			metricsOutput      []string
			metricsOutputError []string
			semanticVersion    semver.Version
			versionString      string
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			nsMap = NamespaceMappings{}
			serverLabels = prometheus.Labels{
				ServerLabelName: "hostname:5432",
			}
			queries = map[string]string{"tablexpto": `
							SELECT  col1,
									col2,
									col3,
									col4,
									col5,
									col6
							FROM tablexpto`,
			}
			metricMaps = map[string]IntermediateMetricMap{
				"tablexpto": {
					map[string]ColumnMapping{
						"col1": {LABEL, "col1", nil},
						"col2": {COUNTER, "col2", nil},
						"col3": {GAUGE, "col3", nil},
						"col4": {MAPPEDMETRIC, "col4", nil},
						"col5": {DURATION, "col5", nil},
						"col6": {HISTOGRAM, "col6", nil},
					},
				},
			}

			metricsOutput = []string{
				`Desc{fqName: "tablexpto_col2", help: "col2", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col3", help: "col3", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col4", help: "col4", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col5_milliseconds", help: "col5", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col2", help: "col2", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col3", help: "col3", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col4", help: "col4", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "tablexpto_col5_milliseconds", help: "col5", constLabels: {server="hostname:5432"}, variableLabels: [col1]}`,
				`Desc{fqName: "pg_static", help: "Version string as reported by postgres", constLabels: {server="hostname:5432"}, variableLabels: [version short_version]}`,
			}

			metricsOutputError = []string{
				`Desc{fqName: "pg_static", help: "Version string as reported by postgres", constLabels: {server="hostname:5432"}, variableLabels: [version short_version]}`,
			}

			semanticVersion = semver.Version{
				Major: 10,
				Minor: 1,
				Patch: 0,
			}
			versionString = "10.1.0"

		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should pass if QueryNamespaceMappings works even with errors", func() {
			db, mock, _ := sqlmock.New()

			defer db.Close()

			rows := sqlmock.NewRows([]string{"col1", "col2", "col3", "col4", "col5", "col6"}).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401})).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401}))

			mock.ExpectQuery(queries["tablexpto"]).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.QueryNamespaceMappings(ch, db, serverLabels, queries, metricMaps, semanticVersion, versionString)
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
			Expect(reflect.DeepEqual(list, metricsOutput)).To(BeTrue())
		})

		It("should pass with namespace found with empty query", func() {
			db, mock, _ := sqlmock.New()

			defer db.Close()

			rows := sqlmock.NewRows([]string{"col1", "col2", "col3", "col4", "col5", "col6"}).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401})).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401}))

			mock.ExpectQuery(queries["dummy"]).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.QueryNamespaceMappings(ch, db, serverLabels, map[string]string{"tablexpto": ""}, metricMaps, semanticVersion, versionString)
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
			Expect(reflect.DeepEqual(list, metricsOutputError)).To(BeTrue())
		})

		It("should pass with namespace and query not found", func() {
			db, mock, _ := sqlmock.New()

			defer db.Close()

			rows := sqlmock.NewRows([]string{"col1", "col2", "col3", "col4", "col5", "col6"}).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401})).
				AddRow("dummy", 1, 1, 1, 1, pq.Array([]float64{235, 401}))

			mock.ExpectQuery(queries["dummy"]).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.QueryNamespaceMappings(ch, db, serverLabels, map[string]string{"notFound": ""}, metricMaps, semanticVersion, versionString)
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
			Expect(reflect.DeepEqual(list, metricsOutputError)).To(BeTrue())
		})

		It("should pass with query error", func() {
			db, mock, _ := sqlmock.New()
			defer db.Close()

			mock.ExpectQuery(queries["tablexpto"]).WillReturnError(errDummy)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.QueryNamespaceMappings(ch, db, serverLabels, queries, metricMaps, semanticVersion, versionString)
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
			Expect(reflect.DeepEqual(list, metricsOutputError)).To(BeTrue())
		})

		It("should pass multiple histogram", func() {
			db, mock, _ := sqlmock.New()

			iqueries := map[string]string{"tablexpto": `
							SELECT  col1,
									col1_bucket,
									col1_sum,
									col1_count									
							FROM tablexpto`,
			}

			mmaps := map[string]IntermediateMetricMap{
				"tablexpto": {
					map[string]ColumnMapping{
						"col1":        {HISTOGRAM, "col1", nil},
						"col1_bucket": {LABEL, "col1_bucket", nil},
						"col1_sum":    {LABEL, "col1_sum", nil},
						"col1_count":  {LABEL, "col1_count", nil},
					},
				},
			}

			defer db.Close()

			rows := sqlmock.NewRows([]string{"col1", "col1_bucket", "col1_sum", "col1_count"}).
				AddRow(pq.Array([]float64{235, 401}), pq.Array([]float64{235, 401}), pq.Array([]float64{235, 401}), 1000)

			mock.ExpectQuery(iqueries["tablexpto"]).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.QueryNamespaceMappings(ch, db, serverLabels, iqueries, mmaps, semanticVersion, versionString)
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
			Expect(reflect.DeepEqual(list, metricsOutputError)).To(BeTrue())
		})

	})

	Context("SetupInternalMetrics", func() {
		var (
			ctrl *gomock.Controller

			metricsOutput []string
			nsMap         NamespaceMappings
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			nsMap = NamespaceMappings{}

			metricsOutput = []string{
				`Desc{fqName: "pg_exporter_last_scrape_duration_seconds", help: "Duration of the last scrape of metrics from PostgresSQL.", constLabels: {}, variableLabels: []}`,
				`Desc{fqName: "pg_exporter_scrapes_total", help: "Total number of times PostgresSQL was scraped for metrics.", constLabels: {}, variableLabels: []}`,
				`Desc{fqName: "pg_rds_current_capacity", help: "Current Aurora capacity units", constLabels: {}, variableLabels: []}`,
				`Desc{fqName: "pg_rds_database_connections", help: "Current Aurora database connections", constLabels: {}, variableLabels: []}`,
			}

		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should pass if SetInternalMetrics works ", func() {
			ch := make(chan prometheus.Metric)
			list := []string{}

			go func() {
				nsMap.SetInternalMetrics(ch, 10, 1, 2, 2)
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

			Expect(reflect.DeepEqual(list, metricsOutput)).To(BeTrue())
		})

	})

})
