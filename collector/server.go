// Copyright 2022 The Prometheus Authors
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

package collector

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type server struct {
	dsn  string
	name string
	db   *sql.DB
}

func makeServer(dsn string) (*server, error) {
	name, err := parseServerName(dsn)
	if err != nil {
		return nil, err
	}
	return &server{
		dsn:  dsn,
		name: name,
	}, nil
}

func (s *server) GetDB() (*sql.DB, error) {
	if s.db != nil {
		return s.db, nil
	}

	db, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s.db = db

	return s.db, nil
}

func (s *server) GetName() string {
	return s.name
}

func (s *server) String() string {
	return s.name
}

func parseServerName(url string) (string, error) {
	dsn, err := pq.ParseURL(url)
	if err != nil {
		dsn = url
	}

	pairs := strings.Split(dsn, " ")
	kv := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		splitted := strings.SplitN(pair, "=", 2)
		if len(splitted) != 2 {
			return "", fmt.Errorf("malformed dsn %q", dsn)
		}
		// Newer versions of pq.ParseURL quote values so trim them off if they exist
		key := strings.Trim(splitted[0], "'\"")
		value := strings.Trim(splitted[1], "'\"")
		kv[key] = value
	}

	var fingerprint string

	if host, ok := kv["host"]; ok {
		fingerprint += host
	} else {
		fingerprint += "localhost"
	}

	if port, ok := kv["port"]; ok {
		fingerprint += ":" + port
	} else {
		fingerprint += ":5432"
	}

	return fingerprint, nil
}
