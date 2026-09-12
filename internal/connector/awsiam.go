// Copyright 2021 The Prometheus Authors
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
	"fmt"
	"net/url"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/lib/pq"
)

// AWSIAMConnector is a Connector that uses AWS IAM authentication to connect to a Postgres database.
// It mints a fresh IAM auth token on every connection attempt, optionally assuming a role via STS if RoleARN is set.
type AWSIAMConnector struct {
	DBEndpoint string // host:port
	DBUser     string
	Database   string
	Region     string
	RoleARN    string            // optional; assumed via STS if set
	Options    map[string]string // extra query parameters, e.g. sslmode

	once  sync.Once
	creds aws.CredentialsProvider
	err   error
}

// NewAWSIAMConnector builds an AWSIAMConnector for the given parameters.
func NewAWSIAMConnector(dbEndpoint, dbUser, database, region, roleARN string, options map[string]string) *AWSIAMConnector {
	return &AWSIAMConnector{
		DBEndpoint: dbEndpoint,
		DBUser:     dbUser,
		Database:   database,
		Region:     region,
		RoleARN:    roleARN,
		Options:    options,
	}
}

// Connect implements the Connector interface.
// It mints a fresh IAM auth token on every connection attempt, optionally assuming a role via STS if RoleARN is set.
func (c *AWSIAMConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c.once.Do(func() { c.creds, c.err = c.loadCredentials(ctx) })
	if c.err != nil {
		return nil, fmt.Errorf("loading AWS credentials: %w", c.err)
	}

	token, err := auth.BuildAuthToken(ctx, c.DBEndpoint, c.Region, c.DBUser, c.creds)
	if err != nil {
		return nil, fmt.Errorf("minting IAM auth token: %w", err)
	}

	connector, err := pq.NewConnector(c.dsn(token))
	if err != nil {
		return nil, err
	}
	return connector.Connect(ctx)
}

// Driver implements the Connector interface.
// It returns the underlying pq.Driver.
func (c *AWSIAMConnector) Driver() driver.Driver {
	return pq.Driver{}
}

// loadCredentials loads AWS credentials, optionally assuming a role if RoleARN is set.
// It caches the credentials for reuse across Connect calls.
func (c *AWSIAMConnector) loadCredentials(ctx context.Context) (aws.CredentialsProvider, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region))
	if err != nil {
		return nil, err
	}
	if c.RoleARN == "" {
		return cfg.Credentials, nil
	}
	return aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), c.RoleARN)), nil
}

// dsn builds the connection string for a single connection attempt, with the freshly minted token as the password.
func (c *AWSIAMConnector) dsn(token string) string {
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(c.DBUser, token),
		Host:   c.DBEndpoint,
		Path:   "/" + c.Database,
	}
	q := u.Query()
	for k, v := range c.Options {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
