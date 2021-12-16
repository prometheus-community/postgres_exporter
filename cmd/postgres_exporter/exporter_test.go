package postgres_exporter_test

import (
	. "github.com/everestsystems/postgres_exporter/cmd/postgres_exporter"
	"github.com/everestsystems/postgres_exporter/cmd/postgres_exporter/mocks"
	. "github.com/everestsystems/postgres_exporter/cmd/postgres_exporter/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Exporter", func() {
	Context("Exporter", func() {
		var (
			ctrl       *gomock.Controller
			e          Exporter
			rdsMetrics *MockRDSMetricsAPI
			server     *MockServerAPI
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			rdsMetrics = mocks.NewMockRDSMetricsAPI(ctrl)
			server = mocks.NewMockServerAPI(ctrl)

			e = Exporter{
				ClusterID:    "dummy",
				RdsMetrics:   rdsMetrics,
				Server:       server,
				TotalScrapes: 0,
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should works if NewExporter works", func() {
			opts := []ExporterOpt{
				ClusterID("dummy"),
				RdsMetrics(rdsMetrics),
				ServerInstance(server),
			}
			exporter := NewExporter(opts...)
			Expect(exporter.ClusterID == e.ClusterID).To(BeTrue())
		})

		It("should works if Scrape works", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(10), nil)

			server.EXPECT().Open().Return(nil)
			server.EXPECT().Scrape(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			server.EXPECT().Close().Return(nil)

			err := e.Scrape(ch)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail if RdsCurrentCapacity fail", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), errDummy)

			err := e.Scrape(ch)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if RdsCurrentConnections fail", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(0), errDummy)

			err := e.Scrape(ch)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if Open fail", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(10), nil)

			server.EXPECT().Open().Return(errDummy)

			err := e.Scrape(ch)
			Expect(err).To(MatchError(errDummy))
		})

		It("should fail if scrape fail", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(10), nil)

			server.EXPECT().Open().Return(nil)
			server.EXPECT().Scrape(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errDummy)

			err := e.Scrape(ch)
			Expect(err).To(MatchError(errDummy))
		})

		It("should fail if close fail", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(2), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(10), nil)

			server.EXPECT().Open().Return(nil)
			server.EXPECT().Scrape(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			server.EXPECT().Close().Return(errDummy)

			err := e.Scrape(ch)
			Expect(err).To(MatchError(errDummy))
		})

		It("should works with aurora serverless sleeping", func() {
			ch := make(chan prometheus.Metric)

			rdsMetrics.EXPECT().RdsCurrentCapacity(e.ClusterID).Return(int64(0), nil)
			rdsMetrics.EXPECT().RdsCurrentConnections(e.ClusterID).Return(int64(0), nil)

			err := e.Scrape(ch)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
