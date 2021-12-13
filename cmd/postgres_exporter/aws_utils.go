package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (e *Exporter) rdsCapacity() (int32, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("error load default config: %w", err)
	}

	stsSvc := sts.NewFromConfig(cfg)

	cfg2, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscredsv2.NewAssumeRoleProvider(
				stsSvc,
				e.iamRoleArn,
			)),
		),
	)
	if err != nil {
		return 0, fmt.Errorf("error assume role: %w", err)
	}

	rdsClient := rds.NewFromConfig(cfg2)

	output, err := rdsClient.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(fmt.Sprintf("tenant-%s", e.tenantID)),
	})
	if err != nil {
		return 0, fmt.Errorf("error describe cluster: %w", err)
	}

	return *output.DBClusters[0].Capacity, nil
}

func (e *Exporter) rdsConnections() (int64, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("error load default config: %w", err)
	}

	stsSvc := sts.NewFromConfig(cfg)

	cfg2, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscredsv2.NewAssumeRoleProvider(
				stsSvc,
				e.iamRoleArn,
			)),
		),
	)
	if err != nil {
		return 0, fmt.Errorf("error assume role: %w", err)
	}

	cloudwatchClient := cloudwatch.NewFromConfig(cfg2)

	output, err := cloudwatchClient.GetMetricStatistics(context.TODO(), &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(time.Now().UTC().Add(time.Second * -60)),
		EndTime:    aws.Time(time.Now().UTC()),
		MetricName: aws.String("DatabaseConnections"),
		Namespace:  aws.String("AWS/RDS"),
		Period:     aws.Int32(60),
		Dimensions: []types.Dimension{{
			Name:  aws.String("DBClusterIdentifier"),
			Value: aws.String(fmt.Sprintf("tenant-%s", e.tenantID)),
		}},
		Statistics: []types.Statistic{"Maximum"},
		Unit:       types.StandardUnitCount,
	})

	if err != nil {
		return 0, fmt.Errorf("error describe cluster: %w", err)
	}

	if len(output.Datapoints) == 0 {
		return 0, nil
	}
	return int64(*output.Datapoints[0].Maximum), nil
}
