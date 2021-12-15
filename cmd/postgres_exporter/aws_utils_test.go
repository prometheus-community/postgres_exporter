package postgres_exporter_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter"
	"github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter/mocks"
)

var errDummy = fmt.Errorf("dummyError")

var _ = Describe("AwsUtils", func() {
	Context("AwsUtils", func() {
		var (
			ctrl           *gomock.Controller
			cloudWatchMock *mocks.MockCloudWatchAPI
			rdsMock        *mocks.MockRDSAPI

			a AwsUtils
		)

		const tenantID = "dummy"
		const iamRoleArn = "dummy"

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			rdsMock = mocks.NewMockRDSAPI(ctrl)
			cloudWatchMock = mocks.NewMockCloudWatchAPI(ctrl)

			a = AwsUtils{
				RdsClient:        rdsMock,
				CloudwatchClient: cloudWatchMock,
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should fail if DescribeDBClusters fails", func() {
			rdsMock.EXPECT().DescribeDBClusters(gomock.Any()).Return(nil, errDummy)

			_, err := a.RdsCurrentCapacity(tenantID)
			Expect(err).To(MatchError(errDummy))
		})

		It("should pass if DescribeDBClusters works", func() {
			rdsMock.EXPECT().DescribeDBClusters(gomock.Any()).Return(&rds.DescribeDBClustersOutput{
				DBClusters: []*rds.DBCluster{&rds.DBCluster{
					Capacity: aws.Int64(2),
				}},
			}, nil)

			_, err := a.RdsCurrentCapacity(tenantID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail if GetMetricStatistics fails", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(nil, errDummy)

			_, err := a.RdsCurrentConnections(tenantID)
			Expect(err).To(MatchError(errDummy))
		})

		It("should pass if GetMetricStatistics returns empty Datapoints", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(&cloudwatch.GetMetricStatisticsOutput{
				Datapoints: []*cloudwatch.Datapoint{},
			}, nil)

			_, err := a.RdsCurrentConnections(tenantID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass if GetMetricStatistics returns Datapoints", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(&cloudwatch.GetMetricStatisticsOutput{
				Datapoints: []*cloudwatch.Datapoint{&cloudwatch.Datapoint{
					Maximum: aws.Float64(0),
				}},
			}, nil)

			_, err := a.RdsCurrentConnections(tenantID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
