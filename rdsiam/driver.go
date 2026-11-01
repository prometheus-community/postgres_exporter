// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package rdsiam provides database/sql drivers that connect to PostgreSQL
// on Amazon RDS or Aurora using IAM database authentication. They accept the
// same DSNs as lib/pq, minus the password: a fresh IAM auth token is minted
// for each new connection and used as the password.
//
// To collect metrics through one, set config.Config.DriverName to the name
// that DriverName returns.
package rdsiam

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/lib/pq"
)

var (
	driversMu sync.Mutex
	drivers   = map[[2]string]string{}
)

// DriverName returns the name of a driver, for sql.Open, that authenticates
// in region as roleARN. Either may be empty to use the AWS SDK's default
// chain (AWS_REGION, profiles, web identity, instance metadata, ...). Each
// distinct pair is registered once and shares its credentials across
// connections.
func DriverName(region, roleARN string) string {
	driversMu.Lock()
	defer driversMu.Unlock()
	key := [2]string{region, roleARN}
	if name, ok := drivers[key]; ok {
		return name
	}
	name := fmt.Sprintf("postgres-rds-iam-%d", len(drivers))
	sql.Register(name, &Driver{region: region, roleARN: roleARN})
	drivers[key] = name
	return name
}

// Test seams.
var (
	loadAWSConfig = loadDefaultAWSConfig
	buildToken    = auth.BuildAuthToken
	connect       = func(ctx context.Context, cfg pq.Config) (driver.Conn, error) {
		c, err := pq.NewConnectorConfig(cfg)
		if err != nil {
			return nil, err
		}
		return c.Connect(ctx)
	}
)

// Driver is a database/sql driver for IAM-authenticated PostgreSQL.
type Driver struct {
	region  string
	roleARN string

	// sem is held while awsCfg is read or loaded. It's a channel rather than
	// a mutex so that a connection waiting for another one's load gives up
	// when its own context is done.
	semOnce sync.Once
	sem     chan struct{}
	awsCfg  *aws.Config
}

var _ driver.DriverContext = (*Driver)(nil)

// Open implements driver.Driver.
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	c, err := d.OpenConnector(dsn)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

// OpenConnector implements driver.DriverContext. sql.Open calls it once per
// DSN, so the DSN is parsed and validated once, not on every connection.
func (d *Driver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := pq.NewConfig(dsn)
	if err != nil {
		return nil, err
	}
	if len(cfg.Multi) > 0 {
		return nil, errors.New("rds iam: multiple hosts are not supported")
	}
	if cfg.Password != "" {
		return nil, errors.New("rds iam: a password must not be set, the IAM auth token is used instead")
	}
	switch cfg.SSLMode {
	case pq.SSLModeDisable, pq.SSLModeAllow, pq.SSLModePrefer:
		return nil, fmt.Errorf("rds iam: sslmode %q could send the auth token in plaintext", cfg.SSLMode)
	}
	return &connector{driver: d, cfg: cfg}, nil
}

type connector struct {
	driver *Driver
	cfg    pq.Config
}

// Connect mints a token for the host, port and user lib/pq will connect with,
// and connects using it as the password.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	awsCfg, err := c.driver.aws(ctx)
	if err != nil {
		return nil, fmt.Errorf("rds iam: loading AWS config: %w", err)
	}
	endpoint := net.JoinHostPort(c.cfg.Host, strconv.Itoa(int(c.cfg.Port)))
	token, err := buildToken(ctx, endpoint, awsCfg.Region, c.cfg.User, awsCfg.Credentials)
	if err != nil {
		return nil, fmt.Errorf("rds iam: building auth token: %w", err)
	}
	cfg := c.cfg.Clone()
	cfg.Password = token
	return connect(ctx, cfg)
}

func (c *connector) Driver() driver.Driver { return c.driver }

// aws loads d's region and credentials once, retrying on the next call if
// loading failed.
func (d *Driver) aws(ctx context.Context) (aws.Config, error) {
	d.semOnce.Do(func() { d.sem = make(chan struct{}, 1) })
	select {
	case d.sem <- struct{}{}:
	case <-ctx.Done():
		return aws.Config{}, ctx.Err()
	}
	defer func() { <-d.sem }()
	if d.awsCfg != nil {
		return *d.awsCfg, nil
	}
	cfg, err := loadAWSConfig(ctx, d.region, d.roleARN)
	if err != nil {
		return aws.Config{}, err
	}
	d.awsCfg = &cfg
	return cfg, nil
}

// loadDefaultAWSConfig loads region and credentials from the AWS SDK's default
// chain, overriding the region if set and assuming roleARN if set. The
// credentials providers it returns are cached by the SDK.
func loadDefaultAWSConfig(ctx context.Context, region, roleARN string) (aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return aws.Config{}, err
	}
	if cfg.Region == "" {
		// Fall back to the instance's region on EC2, in a single attempt:
		// elsewhere the lookup fails on every new connection.
		out, err := imds.New(imds.Options{Retryer: aws.NopRetryer{}}).GetRegion(ctx, nil)
		if err != nil {
			return aws.Config{}, fmt.Errorf("no AWS region configured; set aws_iam.region or AWS_REGION (EC2 instance metadata: %w)", err)
		}
		cfg.Region = out.Region
	}
	if cfg.Region == "" {
		return aws.Config{}, errors.New("no AWS region configured; set aws_iam.region or AWS_REGION")
	}
	if roleARN != "" {
		cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), roleARN))
	}
	return cfg, nil
}
