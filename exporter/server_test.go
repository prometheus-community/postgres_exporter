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
	"log/slog"
	"net"
	"testing"
	"time"
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

func TestExporterCloseServersPreventsNewConnections(t *testing.T) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	e := NewExporter(nil, slog.Default())
	e.CloseServers()
	e.CloseServers()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	server, err := e.servers.GetServer(ctx, "postgres://test@"+listener.Addr().String()+"/postgres?sslmode=disable&connect_timeout=1")
	if server != nil || err == nil {
		t.Fatalf("GetServer() after close returned server = %t, error = %v", server != nil, err)
	}
	if err := listener.SetDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	conn, err := listener.Accept()
	if err == nil {
		conn.Close()
		t.Fatal("GetServer() opened a database connection after CloseServers()")
	}
	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		t.Fatalf("Accept() error = %v, want timeout", err)
	}
}
