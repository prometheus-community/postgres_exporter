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

	"github.com/lib/pq"
)

// Static is a Connector for a fixed DSN, e.g. userpass auth. It connects the
// same way sql.Open("postgres", dsn) does; unlike an auth method that mints
// short-lived credentials, there is nothing to refresh between connections.
type StaticConnector struct {
	connector driver.Connector
}

// NewStaticConnector builds a StaticConnector.
// It returns an error if the DSN is invalid.
func NewStaticConnector(dsn string) (*StaticConnector, error) {
	c, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	return &StaticConnector{connector: c}, nil
}

// Connect implements the Connector interface.
// It returns a new connection to the database.
func (s *StaticConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return s.connector.Connect(ctx)
}

// Driver implements the Connector interface.
// It returns the underlying pq.Driver.
func (s *StaticConnector) Driver() driver.Driver {
	return s.connector.Driver()
}
