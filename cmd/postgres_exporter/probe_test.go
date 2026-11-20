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

package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestHandleProbeRequiresTarget(t *testing.T) {
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	handler := handleProbe(
		promslog.NewNopLogger(),
		authHandler,
		config.NewConfigWithDefaults(),
	)
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
}

// fakePostgres is a minimal PostgreSQL wire-protocol server. While accepting
// is true it completes the startup handshake and answers the exporter's
// connection setup queries; otherwise it refuses the connection the way a
// server at its connection limit does.
type fakePostgres struct {
	listener  net.Listener
	accepting atomic.Bool
}

func newFakePostgres(t *testing.T) *fakePostgres {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	f := &fakePostgres{listener: listener}
	f.accepting.Store(true)
	t.Cleanup(func() { listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go f.serve(conn)
		}
	}()
	return f
}

func (f *fakePostgres) addr() string { return f.listener.Addr().String() }

func pgMessage(kind byte, body []byte) []byte {
	msg := make([]byte, 5, 5+len(body))
	msg[0] = kind
	binary.BigEndian.PutUint32(msg[1:], uint32(4+len(body)))
	return append(msg, body...)
}

func pgErrorResponse(severity, code, message string) []byte {
	return pgMessage('E', []byte("S"+severity+"\x00C"+code+"\x00M"+message+"\x00\x00"))
}

// pgRows builds a RowDescription and one DataRow of text columns.
func pgRows(values ...string) []byte {
	desc := binary.BigEndian.AppendUint16(nil, uint16(len(values)))
	var row []byte
	row = binary.BigEndian.AppendUint16(row, uint16(len(values)))
	for i, value := range values {
		desc = append(desc, fmt.Sprintf("c%d", i)...)
		desc = append(desc, 0)
		desc = binary.BigEndian.AppendUint32(desc, 0)  // table oid
		desc = binary.BigEndian.AppendUint16(desc, 0)  // attribute number
		desc = binary.BigEndian.AppendUint32(desc, 25) // text
		desc = binary.BigEndian.AppendUint16(desc, 0xffff)
		desc = binary.BigEndian.AppendUint32(desc, 0xffffffff)
		desc = binary.BigEndian.AppendUint16(desc, 0) // text format
		row = binary.BigEndian.AppendUint32(row, uint32(len(value)))
		row = append(row, value...)
	}
	out := pgMessage('T', desc)
	out = append(out, pgMessage('D', row)...)
	return append(out, pgMessage('C', []byte("SELECT 1\x00"))...)
}

func (f *fakePostgres) serve(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	var length [4]byte
	if _, err := io.ReadFull(r, length[:]); err != nil {
		return
	}
	if _, err := io.CopyN(io.Discard, r, int64(binary.BigEndian.Uint32(length[:]))-4); err != nil {
		return
	}
	if !f.accepting.Load() {
		conn.Write(pgErrorResponse("FATAL", "53300", "sorry, too many clients already"))
		return
	}
	conn.Write(pgMessage('R', []byte{0, 0, 0, 0}))
	ready := pgMessage('Z', []byte("I"))
	conn.Write(ready)
	for {
		var header [5]byte
		if _, err := io.ReadFull(r, header[:]); err != nil {
			return
		}
		body := make([]byte, binary.BigEndian.Uint32(header[1:])-4)
		if _, err := io.ReadFull(r, body); err != nil {
			return
		}
		if header[0] != 'Q' {
			return
		}
		query := string(body)
		var reply []byte
		switch {
		case strings.Contains(query, "SELECT version"):
			reply = pgRows("PostgreSQL 17.0 on x86_64-pc-linux-gnu")
		case strings.Contains(query, "current_database()"):
			reply = pgRows("appdb", "f")
		default:
			reply = pgErrorResponse("ERROR", "42501", "permission denied")
		}
		conn.Write(append(reply, ready...))
	}
}

func probeLines(t *testing.T, handler http.Handler, target string, prefix string) []string {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(target), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if got, want := response.Code, http.StatusOK; got != want {
		t.Fatalf("status code = %d, want %d: %s", got, want, response.Body)
	}
	var lines []string
	for _, line := range strings.Split(response.Body.String(), "\n") {
		if strings.HasPrefix(line, prefix) {
			lines = append(lines, line)
		}
	}
	return lines
}

// TestHandleProbeReportsConnectionHealth checks that every /probe request
// reports pg_scrape_connection_success for its target, with the same labels
// while the connection fails and after it recovers.
func TestHandleProbeReportsConnectionHealth(t *testing.T) {
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	handler := handleProbe(promslog.NewNopLogger(), authHandler, config.NewConfigWithDefaults())

	server := newFakePostgres(t)
	target := "postgresql://exporter@" + server.addr() + "/appdb?sslmode=disable&connect_timeout=5"
	series := func(value int) []string {
		return []string{fmt.Sprintf(`pg_scrape_connection_success{datname="appdb",server=%q} %d`, server.addr(), value)}
	}

	server.accepting.Store(false)
	if got := probeLines(t, handler, target, "pg_scrape_connection_success"); !slices.Equal(got, series(0)) {
		t.Fatalf("while refusing: got %q, want %q", got, series(0))
	}
	if got := probeLines(t, handler, target, "pg_scrape_collector_success"); len(got) != 0 {
		t.Fatalf("while refusing, per-collector success was reported: %q", got)
	}

	server.accepting.Store(true)
	if got := probeLines(t, handler, target, "pg_scrape_connection_success"); !slices.Equal(got, series(1)) {
		t.Fatalf("after recovery: got %q, want %q", got, series(1))
	}
	if got := probeLines(t, handler, target, "pg_scrape_collector_success"); len(got) == 0 {
		t.Fatal("after recovery, no per-collector success was reported")
	}
}
