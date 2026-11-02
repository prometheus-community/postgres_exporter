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
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"syscall"
	"testing"
	"time"

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

func TestProbeConnectionCleanup(t *testing.T) {
	for _, stage := range []string{"startup", "ssl", "query", "connect_timeout", "connect_timeout_expired"} {
		t.Run(stage, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			var mu sync.Mutex
			var connections []net.Conn
			blocked := make(chan struct{}, 2)
			closed := make(chan struct{}, 2)
			readErrors := make(chan error, 2)
			cancelRequests := make(chan struct{}, 2)
			go func() {
				for {
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					mu.Lock()
					connections = append(connections, conn)
					mu.Unlock()
					go func() {
						defer conn.Close()
						var size uint32
						if binary.Read(conn, binary.BigEndian, &size) != nil {
							return
						}
						if size < 8 || size > 1<<20 {
							return
						}
						startup := make([]byte, size-4)
						if _, err := io.ReadFull(conn, startup); err != nil {
							return
						}
						if binary.BigEndian.Uint32(startup) == 80877102 {
							cancelRequests <- struct{}{}
							return
						}
						defer func() { closed <- struct{}{} }()
						if stage == "ssl" && binary.BigEndian.Uint32(startup) != 80877103 {
							t.Error("expected an SSL negotiation request")
						}
						if stage == "query" || stage == "connect_timeout" {
							// AuthenticationOk, BackendKeyData, ReadyForQuery;
							// withhold the response to the first SQL command.
							_, err := conn.Write([]byte{
								'R', 0, 0, 0, 8, 0, 0, 0, 0,
								'K', 0, 0, 0, 12, 0, 0, 0, 1, 0, 0, 0, 2,
								'Z', 0, 0, 0, 5, 'I',
							})
							if err != nil {
								return
							}
							var command [1]byte
							if _, err := io.ReadFull(conn, command[:]); err != nil {
								return
							}
						}
						blocked <- struct{}{}
						_, err := io.Copy(io.Discard, conn)
						readErrors <- err
					}()
				}
			}()

			finished := make(chan struct{})
			authHandler, err := config.NewHandler(prometheus.NewRegistry())
			if err != nil {
				t.Fatal(err)
			}
			probe := handleProbe(promslog.NewNopLogger(), authHandler, config.NewConfigWithDefaults())
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(finished)
				probe(w, r)
			}))
			t.Cleanup(func() {
				listener.Close()
				mu.Lock()
				for _, conn := range connections {
					conn.Close()
				}
				mu.Unlock()
				server.Close()
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sslmode := "disable"
			if stage == "ssl" {
				sslmode = "require"
			}
			target := fmt.Sprintf("postgresql://test@%s/test?sslmode=%s", listener.Addr(), sslmode)
			if stage == "connect_timeout" {
				target += "&connect_timeout=30"
			} else if stage == "connect_timeout_expired" {
				target += "&connect_timeout=1"
			}
			request, err := http.NewRequestWithContext(ctx, http.MethodGet,
				server.URL+"/probe?target="+url.QueryEscape(target), nil)
			if err != nil {
				t.Fatal(err)
			}
			go func() {
				response, _ := server.Client().Do(request)
				if response != nil {
					response.Body.Close()
				}
			}()

			await := func(ch <-chan struct{}, description string) {
				t.Helper()
				select {
				case <-ch:
				case <-time.After(5 * time.Second):
					t.Fatalf("timed out waiting for %s", description)
				}
			}
			// Both the legacy exporter and the new collectors must be blocked.
			await(blocked, "first database connection")
			await(blocked, "second database connection")
			if stage == "connect_timeout" {
				// pq has canceled its temporary connect_timeout context.
				// This must not close a connection still used by the scrape.
				select {
				case <-closed:
					t.Fatal("connect_timeout closed an established connection")
				case <-time.After(2 * time.Second):
				}
			}
			if stage != "connect_timeout_expired" {
				cancel()
			}
			if stage == "query" || stage == "connect_timeout" {
				await(cancelRequests, "first SQL cancellation")
				await(cancelRequests, "second SQL cancellation")
			}
			await(finished, "probe completion")
			await(closed, "first database disconnect")
			await(closed, "second database disconnect")
			for range 2 {
				if err := <-readErrors; !errors.Is(err, syscall.ECONNRESET) {
					t.Errorf("expected a TCP reset, got %v", err)
				}
			}
		})
	}
}
