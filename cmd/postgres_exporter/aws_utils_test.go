package postgres_exporter_test

import (
	"fmt"

	. "."
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	. "github.com/golang/mock/mockgen/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thiagosantosleite/postgres_exporter/cmd/tests/mocks"
)

var errDummy = fmt.Errorf("dummyError")

var _ = Describe("AwsUtils", func() {
	Context("AwsUtils", func() {
		var (
			ctrl           *gomock.Controller
			cloudWatchMock *mocks.MockCloudWatchAPI
			rdsMock        *mocks.MockRDSAPI
		)

		const tenantID = "dummy"
		const iamRoleArn = "dummy"

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			rdsMock = mocks.NewMockRDSAPI(ctrl)
			cloudWatchMock = mocks.NewMockCloudWatchAPI(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should fail if DescribeDBClusters fails", func() {
			rdsMock.EXPECT().DescribeDBClusters(gomock.Any()).Return(nil, errDummy)

			_, err := RdsCurrentCapacity(tenantID, rdsMock)
			Expect(err).To(MatchError(errDummy))
		})

		It("should pass if DescribeDBClusters works", func() {
			rdsMock.EXPECT().DescribeDBClusters(gomock.Any()).Return(&rds.DescribeDBClustersOutput{
				DBClusters: []*rds.DBCluster{&rds.DBCluster{
					Capacity: aws.Int64(2),
				}},
			}, nil)

			_, err := RdsCurrentCapacity(tenantID, rdsMock)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail if GetMetricStatistics fails", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(nil, errDummy)

			_, err := RdsCurrentConnections(tenantID, cloudWatchMock)
			Expect(err).To(MatchError(errDummy))
		})

		It("should pass if GetMetricStatistics returns empty Datapoints", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(&cloudwatch.GetMetricStatisticsOutput{
				Datapoints: []*cloudwatch.Datapoint{},
			}, nil)

			_, err := RdsCurrentConnections(tenantID, cloudWatchMock)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass if GetMetricStatistics returns Datapoints", func() {
			cloudWatchMock.EXPECT().GetMetricStatistics(gomock.Any()).Return(&cloudwatch.GetMetricStatisticsOutput{
				Datapoints: []*cloudwatch.Datapoint{&cloudwatch.Datapoint{
					Maximum: aws.Float64(0),
				}},
			}, nil)

			_, err := RdsCurrentConnections(tenantID, cloudWatchMock)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
