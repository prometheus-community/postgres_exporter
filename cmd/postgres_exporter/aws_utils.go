package main

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

func RdsCurrentCapacity(tenantID string, rdsClient rdsiface.RDSAPI) (int64, error) {
	output, err := rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(getTenant(tenantID)),
	})
	if err != nil {
		return 0, fmt.Errorf("error describe cluster: %w", err)
	}

	return *output.DBClusters[0].Capacity, nil
}

func RdsCurrentConnections(tenantID string, cloudwatchClient cloudwatchiface.CloudWatchAPI) (int64, error) {
	output, err := cloudwatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(time.Now().UTC().Add(time.Second * -60)),
		EndTime:    aws.Time(time.Now().UTC()),
		MetricName: aws.String("DatabaseConnections"),
		Namespace:  aws.String("AWS/RDS"),
		Period:     aws.Int64(60),
		Dimensions: []*cloudwatch.Dimension{{
			Name:  aws.String("DBClusterIdentifier"),
			Value: aws.String(getTenant(tenantID)),
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

func getTenant(tenantID string) string {
	return fmt.Sprintf("tenant-%s", tenantID)
}
