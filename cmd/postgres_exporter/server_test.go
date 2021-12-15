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
	"github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter/mocks"
)

var _ = Describe("Server", func() {
	Context("Close", func() {
		var (
			ctrl *gomock.Controller
			s    Server

			nsMap       *mocks.MockNamespaceMetricsAPI
			settMetrics *mocks.MockSettingsMetricsAPI
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			nsMap = mocks.NewMockNamespaceMetricsAPI(ctrl)
			settMetrics = mocks.NewMockSettingsMetricsAPI(ctrl)

			s = Server{
				Dsn:         "postgresql://user:pass@host:5432/postgres",
				NsMap:       nsMap,
				SettMetrics: settMetrics,
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should pass Close works", func() {
			db, _, err := sqlmock.New()
			s.Db = db

			s.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass if Query works", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db
			r := sqlmock.NewRows([]string{"col1"}).
				AddRow("dummy")

			mock.ExpectQuery("SELECT col1 FROM table").WillReturnRows(r)
			rows, _ := s.Query("SELECT col1 FROM table")
			rows.Next()

			var name string
			rows.Scan(&name)

			Expect(reflect.DeepEqual("dummy", name)).To(BeTrue())
		})

		It("should fail if Query fails", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db

			mock.ExpectQuery("SELECT col1 FROM table").WillReturnError(errDummy)
			_, err := s.Query("SELECT col1 FROM table")

			Expect(err).To(MatchError(errDummy))
		})

		It("should works if Scrape works", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db
			r := sqlmock.NewRows([]string{"version"}).
				AddRow("PostgreSQL 10.14 on x86_64-pc-linux-gnu, compiled by x86_64-unknown-linux-gnu-gcc (GCC) 4.9.4, 64-bit")

			ch := make(chan prometheus.Metric)

			settMetrics.EXPECT().QuerySettings(gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().QueryNamespaceMappings(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().SetInternalMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			mock.ExpectQuery(regexp.QuoteMeta("SELECT version();")).WillReturnRows(r)

			err := s.Scrape(ch, 1000, 1000, 2)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should works when aurora serverless is sleeping", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db
			r := sqlmock.NewRows([]string{"version"}).
				AddRow("PostgreSQL 10.14 on x86_64-pc-linux-gnu, compiled by x86_64-unknown-linux-gnu-gcc (GCC) 4.9.4, 64-bit")

			ch := make(chan prometheus.Metric)

			settMetrics.EXPECT().QuerySettings(gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().QueryNamespaceMappings(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().SetInternalMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			mock.ExpectQuery(regexp.QuoteMeta("SELECT version();")).WillReturnRows(r)

			err := s.Scrape(ch, 0, 0, 0)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fails with invalid postgres version", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db
			r := sqlmock.NewRows([]string{"version"}).
				AddRow("10.14")

			ch := make(chan prometheus.Metric)

			settMetrics.EXPECT().QuerySettings(gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().QueryNamespaceMappings(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().SetInternalMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			mock.ExpectQuery(regexp.QuoteMeta("SELECT version();")).WillReturnRows(r)

			err := s.Scrape(ch, 1000, 1000, 2)
			Expect(err).To(HaveOccurred())
		})

		It("should fails with invalid postgres version", func() {
			db, mock, _ := sqlmock.New()
			s.Db = db

			ch := make(chan prometheus.Metric)

			settMetrics.EXPECT().QuerySettings(gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().QueryNamespaceMappings(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			nsMap.EXPECT().SetInternalMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			mock.ExpectQuery(regexp.QuoteMeta("SELECT version();")).WillReturnError(errDummy)

			err := s.Scrape(ch, 1000, 1000, 2)
			Expect(err).To(HaveOccurred())
		})
	})
})
