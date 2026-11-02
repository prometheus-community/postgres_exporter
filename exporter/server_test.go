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

package exporter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestGetServerRetryHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	started := time.Now()
	_, err := NewServers().GetServer(ctx, "invalid dsn")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetServer() error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed >= time.Second {
		t.Fatalf("GetServer() returned after %v, want cancellation before retry delay completed", elapsed)
	}
}

func TestProbeStopsBeforeNextDatabase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(
		func(expected, actual string) error {
			if err := sqlmock.QueryMatcherEqual.Match(expected, actual); err != nil {
				return err
			}
			cancel()
			return nil
		})))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT version();").WillReturnError(context.Canceled)
	mock.ExpectClose()

	first := "postgresql://test@127.0.0.1:1/first?sslmode=disable"
	second := "postgresql://test@127.0.0.1:1/second?sslmode=disable"
	e := NewExporter([]string{first, second}, promslog.NewNopLogger(), WithContext(ctx))
	defer e.servers.Close()
	e.servers.servers[first] = &Server{ctx: ctx, db: db, logger: promslog.NewNopLogger()}
	e.servers.opts = append(e.servers.opts, func(*Server) {
		t.Error("started another database after the probe was canceled")
	})
	e.scrape(ctx, make(chan prometheus.Metric, 100))
	e.servers.Close()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
