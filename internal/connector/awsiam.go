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

	mu      sync.Mutex
	creds   aws.CredentialsProvider
	loading chan struct{} // non-nil, and closed on completion, while a goroutine is loading creds

	// Test seams; nil means the production implementation.
	loadCreds func(ctx context.Context, region string) (aws.CredentialsProvider, string, error)
	open      func(ctx context.Context, cfg pq.Config) (driver.Conn, error)
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
	creds, region, err := c.getCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS credentials: %w", err)
	}

	token, err := auth.BuildAuthToken(ctx, c.DBEndpoint, region, c.DBUser, creds)
	if err != nil {
		return nil, fmt.Errorf("minting IAM auth token: %w", err)
	}

	cfg, err := pq.NewConfig(c.DSN(token))
	if err != nil {
		return nil, err
	}
	// lib/pq applies a connection service (service= or PGSERVICE) after the
	// DSN, so it could replace the signed host, port and user, the minted
	// token, or the sslmode and send the token somewhere it wasn't meant for.
	if cfg.Service != "" {
		return nil, errors.New("connection services (service= / PGSERVICE) are not supported with IAM authentication")
	}
	if c.open != nil {
		return c.open(ctx, cfg)
	}
	conn, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		return nil, err
	}
	return conn.Connect(ctx)
}

// Driver implements the Connector interface.
// It returns the underlying pq.Driver.
func (c *AWSIAMConnector) Driver() driver.Driver {
	return pq.Driver{}
}

// getCredentials returns the cached credentials, loading them on the first successful call.
// A failed load is not cached, so the next Connect attempt retries it. Only one
// caller loads at a time; the others wait for it but return as soon as their own
// ctx is done, so one slow load can't hold past another caller's timeout.
func (c *AWSIAMConnector) getCredentials(ctx context.Context) (aws.CredentialsProvider, string, error) {
	for {
		c.mu.Lock()
		if c.creds != nil {
			creds, region := c.creds, c.Region
			c.mu.Unlock()
			return creds, region, nil
		}
		if wait := c.loading; wait != nil {
			c.mu.Unlock()
			select {
			case <-wait:
				continue // the load finished; re-check, and retry it if it failed
			case <-ctx.Done():
				return nil, "", ctx.Err()
			}
		}
		done := make(chan struct{})
		c.loading = done
		region := c.Region
		c.mu.Unlock()

		load := c.loadCredentials
		if c.loadCreds != nil {
			load = c.loadCreds
		}
		creds, region, err := load(ctx, region)

		c.mu.Lock()
		c.loading = nil
		if err == nil {
			c.creds = creds
			c.Region = region
		}
		c.mu.Unlock()
		close(done)
		if err != nil {
			return nil, "", err
		}
		return creds, region, nil
	}
}

// loadCredentials loads AWS credentials, optionally assuming a role if RoleARN is set.
func (c *AWSIAMConnector) loadCredentials(ctx context.Context, region string) (aws.CredentialsProvider, string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region), awsconfig.WithEC2IMDSRegion())
	if err != nil {
		return nil, "", err
	}
	if region == "" {
		if cfg.Region == "" {
			return nil, "", errors.New("no AWS region configured; set iam.region (or AWS_REGION)")
		}
		region = cfg.Region
	}
	if c.RoleARN == "" {
		return cfg.Credentials, region, nil
	}
	return aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), c.RoleARN)), region, nil
}
