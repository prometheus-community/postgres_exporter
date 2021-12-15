package postgres_exporter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type AwsUtils struct {
	RdsClient        rdsiface.RDSAPI
	CloudwatchClient cloudwatchiface.CloudWatchAPI
}

// compile-time check that type implements interface.
var _ RDSMetricsAPI = (*AwsUtils)(nil)

func (a *AwsUtils) RdsCurrentCapacity(tenantID string) (int64, error) {
	output, err := a.RdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(GetTenant(tenantID)),
	})
	if err != nil {
		return 0, fmt.Errorf("error describe cluster: %w", err)
	}

	return *output.DBClusters[0].Capacity, nil
}

func (a *AwsUtils) RdsCurrentConnections(tenantID string) (int64, error) {
	output, err := a.CloudwatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(time.Now().UTC().Add(time.Second * -60)),
		EndTime:    aws.Time(time.Now().UTC()),
		MetricName: aws.String("DatabaseConnections"),
		Namespace:  aws.String("AWS/RDS"),
		Period:     aws.Int64(60),
		Dimensions: []*cloudwatch.Dimension{{
			Name:  aws.String("DBClusterIdentifier"),
			Value: aws.String(GetTenant(tenantID)),
		}},
		Statistics: []*string{aws.String(cloudwatch.StatisticMaximum)},
		Unit:       aws.String(cloudwatch.StandardUnitCount),
	})

	if err != nil {
		return 0, fmt.Errorf("error describe cluster: %w", err)
	}

	if len(output.Datapoints) == 0 {
		return 0, nil
	}
	return int64(*output.Datapoints[0].Maximum), nil
}

func NewAWSSession(iamRoleArn string) (*session.Session, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	sessRole, err := session.NewSession(&aws.Config{
		Credentials: stscreds.NewCredentials(
			sess,
			iamRoleArn,
			func(provider *stscreds.AssumeRoleProvider) {
				provider.RoleSessionName = "postgres-exporter"
			},
		),
		Region: sess.Config.Region,
	})
	if err != nil {
		return nil, err
	}
	return sessRole, nil
}
