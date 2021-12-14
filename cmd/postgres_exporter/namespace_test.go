package postgres_exporter_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"

	//. "github.com/onsi/gomega"
	. "github.com/golang/mock/mockgen/model"
)

var _ = Describe("namespace", func() {
	Context("QueryNamespaceMappings", func() {
		var (
			ctrl *gomock.Controller
			//ch   chan<- prometheus.Metric
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			//			serverLabels = "fingerprint=hostname:5432"
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should fail if DescribeDBClusters fails", func() {
			/*			&Server{
							db: db,
							labels: prometheus.Labels{
								serverLabelName: fingerprint,
							},
						}

						err := QueryNamespaceMappings(ch)
						Expect(err).To(MatchError(errDummy))
			*/
		})

	})
})
