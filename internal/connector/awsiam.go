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

package connector

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/lib/pq"
)

// defaultPostgresPort is assumed when a target's DSN omits the port, matching
// lib/pq's own default so IAM behaves the same as userpass for a bare host.
const defaultPostgresPort = "5432"

// AWSIAMConnector implements the database/sql/driver.Connector interface, so it can be used with sql.OpenDB.
var _ driver.Connector = (*AWSIAMConnector)(nil)

// AWSIAMConnector is a Connector that uses AWS IAM authentication to connect to a Postgres database.
// It mints a fresh IAM auth token on every connection attempt, optionally assuming a role via STS if RoleARN is set.
type AWSIAMConnector struct {
	DBEndpoint string // host:port
	DBUser     string
	Region     string
	RoleARN    string // optional; assumed via STS if set
	DSN        func(token string) string

	mu    sync.Mutex
	creds aws.CredentialsProvider
}

// NewAWSIAMConnector builds an AWSIAMConnector for the given parameters.
// dbEndpoint's port defaults to defaultPostgresPort when omitted, since BuildAuthToken requires host:port.
// dsn builds the full connection string for a single connection attempt, with the freshly minted token as the password.
func NewAWSIAMConnector(dbEndpoint, dbUser, region, roleARN string, dsn func(token string) string) (*AWSIAMConnector, error) {
	dbEndpoint, err := withDefaultPort(dbEndpoint)
	if err != nil {
		return nil, err
	}
	return &AWSIAMConnector{
		DBEndpoint: dbEndpoint,
		DBUser:     dbUser,
		Region:     region,
		RoleARN:    roleARN,
		DSN:        dsn,
	}, nil
}

// withDefaultPort returns endpoint with a port, defaulting to defaultPostgresPort
// when none is present. It returns an error if endpoint is otherwise malformed.
func withDefaultPort(endpoint string) (string, error) {
	if _, _, err := net.SplitHostPort(endpoint); err == nil {
		return endpoint, nil
	}
	host, _, err := net.SplitHostPort(endpoint + ":0")
	if err != nil {
		return "", fmt.Errorf("invalid database endpoint %q: %w", endpoint, err)
	}
	return net.JoinHostPort(host, defaultPostgresPort), nil
}

// Connect implements the Connector interface.
// It mints a fresh IAM auth token on every connection attempt, optionally assuming a role via STS if RoleARN is set.
func (c *AWSIAMConnector) Connect(ctx context.Context) (driver.Conn, error) {
	creds, err := c.getCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS credentials: %w", err)
	}

	token, err := auth.BuildAuthToken(ctx, c.DBEndpoint, c.Region, c.DBUser, creds)
	if err != nil {
		return nil, fmt.Errorf("minting IAM auth token: %w", err)
	}

	static, err := NewStaticConnector(c.DSN(token))
	if err != nil {
		return nil, err
	}
	return static.Connect(ctx)
}

// Driver implements the Connector interface.
// It returns the underlying pq.Driver.
func (c *AWSIAMConnector) Driver() driver.Driver {
	return pq.Driver{}
}

// getCredentials returns the cached credentials, loading them on the first successful call.
// A failed load is not cached, so the next Connect attempt retries it.
func (c *AWSIAMConnector) getCredentials(ctx context.Context) (aws.CredentialsProvider, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.creds != nil {
		return c.creds, nil
	}
	creds, err := c.loadCredentials(ctx)
	if err != nil {
		return nil, err
	}
	c.creds = creds
	return c.creds, nil
}

// loadCredentials loads AWS credentials, optionally assuming a role if RoleARN is set.
func (c *AWSIAMConnector) loadCredentials(ctx context.Context) (aws.CredentialsProvider, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region))
	if err != nil {
		return nil, err
	}
	if c.Region == "" {
		if cfg.Region == "" {
			return nil, errors.New("no AWS region configured; set iam.region (or AWS_REGION)")
		}
		c.Region = cfg.Region
	}
	if c.RoleARN == "" {
		return cfg.Credentials, nil
	}
	return aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), c.RoleARN)), nil
}
