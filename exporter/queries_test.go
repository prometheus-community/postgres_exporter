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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/common/promslog"
)

func TestQueryDatabasesDiscardsResultsOnRowError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating mock database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	rows := sqlmock.NewRows([]string{"datname"}).
		AddRow("first").
		AddRow("second").
		RowError(1, errors.New("row iteration failed"))
	mock.ExpectQuery("SELECT datname FROM pg_database").WillReturnRows(rows)
	server := &Server{
		db:     db,
		logger: promslog.NewNopLogger(),
	}

	databases, err := queryDatabases(context.Background(), server)
	if err == nil {
		t.Fatal("querying databases returned nil error after row iteration failed")
	}
	if databases != nil {
		t.Errorf("got databases %v after row iteration failed, want nil", databases)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}
