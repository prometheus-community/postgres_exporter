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
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

// Exercise the real entrypoint in a subprocess so Kingpin flags, signal
// handlers, and os.Exit behave as they do in the standalone exporter.
func TestBootstrapProcess(t *testing.T) {
	if os.Getenv("POSTGRES_EXPORTER_TEST_PROCESS") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"postgres_exporter"}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	os.Exit(2)
}

func exporterCommand(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestBootstrapProcess$", "--"}, args...)...)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "PG") && !strings.HasPrefix(entry, "DATA_SOURCE_") && !strings.HasPrefix(entry, "POSTGRES_EXPORTER_TEST_PROCESS=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "POSTGRES_EXPORTER_TEST_PROCESS=1")
	return cmd
}

func TestBootstrapEarlyExit(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
		fail bool
	}{
		{name: "version", args: []string{"--version"}, want: "postgres_exporter, version"},
		{name: "help", args: []string{"-h"}, want: "--web.max-requests"},
		{name: "dumpmaps", args: []string{"--dumpmaps", "--datasource.uri-file=/missing"}, want: "confl_tablespace"},
		{name: "parse error", args: []string{"--unknown-flag"}, want: "unknown", fail: true},
		{name: "negative limit", args: []string{"--web.max-requests=-1"}, want: "greater than or equal to zero", fail: true},
		{name: "empty path", args: []string{"--web.telemetry-path="}, want: "metrics path must not be empty", fail: true},
		{name: "reserved path", args: []string{"--config.file=", "--web.telemetry-path=/probe"}, want: "reserved for the metrics handler", fail: true},
		{name: "listener error", args: []string{"--config.file=", "--web.listen-address=127.0.0.1:invalid"}, want: "Error running exporter", fail: true},
		{name: "datasource error", args: []string{"--config.file=", "--datasource.uri-file=/missing"}, want: "failed reading data sources", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := exporterCommand(t, tc.args...).CombinedOutput()
			if (err != nil) != tc.fail || !strings.Contains(string(output), tc.want) {
				t.Fatalf("error = %v, output = %s; want failure %v containing %q", err, output, tc.fail, tc.want)
			}
		})
	}
}

func startTestExporter(t *testing.T, env []string, args ...string) (string, *exec.Cmd) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	args = append([]string{"--config.file=", "--web.listen-address=" + address}, args...)
	cmd := exporterCommand(t, args...)
	cmd.Env = append(cmd.Env, env...)
	logs, err := os.Create(filepath.Join(t.TempDir(), "exporter.log"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { logs.Close() })
	cmd.Stdout, cmd.Stderr = logs, logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	baseURL := "http://" + address
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/debug/pprof/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return baseURL, cmd
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	output, _ := os.ReadFile(logs.Name())
	t.Fatalf("exporter did not start: %s", output)
	return "", nil
}

func requestExporter(t *testing.T, method, url string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}

func TestBootstrapMetricsAndRoutes(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		name := "self metrics enabled"
		args := []string{"--web.telemetry-path=/custom"}
		if disabled {
			name = "self metrics disabled"
			args = append(args, "--web.disable-exporter-metrics")
		}
		t.Run(name, func(t *testing.T) {
			baseURL, cmd := startTestExporter(t, nil, args...)
			status, body := requestExporter(t, http.MethodGet, baseURL+"/custom")
			if status != http.StatusOK {
				t.Fatalf("metrics status = %d: %s", status, body)
			}
			parser := expfmt.NewTextParser(model.LegacyValidation)
			metrics, err := parser.TextToMetricFamilies(strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"go_goroutines", "process_start_time_seconds", "postgres_exporter_build_info"} {
				if (metrics[name] != nil) == disabled {
					t.Errorf("%s present = %v, self metrics disabled = %v", name, metrics[name] != nil, disabled)
				}
			}
			reload := metrics["postgres_exporter_config_last_reload_successful"]
			if reload == nil || reload.GetType().String() != "GAUGE" || len(reload.Metric) != 1 || reload.Metric[0].GetGauge().GetValue() != 1 {
				t.Fatalf("unexpected configuration reload status: %v", reload)
			}
			up := metrics["pg_up"]
			if up == nil || up.GetType().String() != "GAUGE" || len(up.Metric) != 1 || len(up.Metric[0].Label) != 0 || up.Metric[0].GetGauge().GetValue() != 0 {
				t.Fatalf("unexpected pg_up: %v", up)
			}
			if status, body := requestExporter(t, http.MethodGet, baseURL+"/"); status != http.StatusOK || !strings.Contains(body, "/custom") {
				t.Fatalf("landing page status = %d: %s", status, body)
			}
			if status, _ := requestExporter(t, http.MethodGet, baseURL+"/probe"); status != http.StatusBadRequest {
				t.Fatalf("probe status = %d", status)
			}
			if status, _ := requestExporter(t, http.MethodPost, baseURL+"/-/reload"); status != http.StatusOK {
				t.Fatalf("reload status = %d", status)
			}
			if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
				t.Fatal(err)
			}
			if status, _ := requestExporter(t, http.MethodGet, baseURL+"/debug/vars"); status != http.StatusOK {
				t.Fatalf("expvar status after SIGHUP = %d", status)
			}
		})
	}
}

func TestBootstrapTelemetryPathEnvironment(t *testing.T) {
	for _, cliPath := range []string{"", "/explicit", "/"} {
		t.Run("CLI path "+cliPath, func(t *testing.T) {
			args := []string{}
			want := "/legacy"
			if cliPath != "" {
				args = append(args, "--web.telemetry-path="+cliPath)
				want = cliPath
			}
			baseURL, _ := startTestExporter(t, []string{webTelemetryPathEnvar + "=/legacy"}, args...)
			status, body := requestExporter(t, http.MethodGet, baseURL+want)
			if status != http.StatusOK || !strings.Contains(body, "# TYPE pg_up gauge") {
				t.Fatalf("metrics at %s: status = %d, body = %s", want, status, body)
			}
			if cliPath != "" && cliPath != "/" {
				_, body := requestExporter(t, http.MethodGet, baseURL+"/legacy")
				if strings.Contains(body, "# TYPE pg_up gauge") {
					t.Fatal("environment variable added an unexpected metrics alias")
				}
			}
		})
	}
}

func TestBootstrapRequestLimits(t *testing.T) {
	for _, tc := range []struct {
		endpoint string
		limit    string
	}{
		{endpoint: "/metrics", limit: "1"},
		{endpoint: "/probe", limit: "1"},
		{endpoint: "/probe", limit: "0"},
	} {
		endpoint, limit := tc.endpoint, tc.limit
		t.Run(endpoint+" limit="+limit, func(t *testing.T) {
			// Hold database connections in startup so the HTTP scrape remains
			// in flight. No PostgreSQL query semantics are involved here.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			databaseAddress := listener.Addr().String()
			entered := make(chan struct{}, 8)
			release := make(chan struct{})
			done := make(chan struct{})
			go func() {
				defer close(done)
				var connections []net.Conn
				defer func() {
					<-release
					for _, conn := range connections {
						conn.Close()
					}
				}()
				for {
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					connections = append(connections, conn)
					entered <- struct{}{}
				}
			}()
			t.Cleanup(func() {
				listener.Close()
				close(release)
				<-done
			})
			baseURL, _ := startTestExporter(t, nil,
				"--web.max-requests="+limit,
				"--datasource.uri="+databaseAddress+"/postgres?sslmode=disable",
			)
			url := baseURL + endpoint
			if endpoint == "/probe" {
				url += "?target=" + databaseAddress + "/postgres?sslmode=disable"
			}
			ctx, cancel := context.WithCancel(context.Background())
			completed := make(chan struct{})
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			go func() {
				defer close(completed)
				resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}()
			t.Cleanup(func() {
				cancel()
				<-completed
			})
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("scrape did not reach the database")
			}
			if limit == "1" {
				status, body := requestExporter(t, http.MethodGet, url)
				if status != http.StatusServiceUnavailable || !strings.Contains(body, "Limit of concurrent requests reached") {
					t.Fatalf("saturated endpoint status = %d: %s", status, body)
				}
			} else {
				if status, _ := requestExporter(t, http.MethodGet, baseURL+"/probe"); status != http.StatusBadRequest {
					t.Fatalf("unlimited probe was rejected before target validation: status = %d", status)
				}
			}
			if status, _ := requestExporter(t, http.MethodPost, baseURL+"/-/reload"); status != http.StatusOK {
				t.Fatalf("reload while scraping: status = %d", status)
			}
			if endpoint == "/probe" {
				if status, _ := requestExporter(t, http.MethodGet, baseURL+"/probe"); limit == "1" && status != http.StatusServiceUnavailable {
					t.Fatalf("probe validation ran ahead of limiter: status = %d", status)
				}
			}
		})
	}
}
