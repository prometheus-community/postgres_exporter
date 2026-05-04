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

package config

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestLoadAuthConfig(t *testing.T) {
	config, err := LoadAuthConfig("testdata/config-good.yaml")
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}
	if len(config.AuthModules) == 0 {
		t.Fatal("LoadAuthConfig() loaded no auth modules")
	}
}

func TestDecodeAuthConfig(t *testing.T) {
	config, err := DecodeAuthConfig(strings.NewReader(`
auth_modules:
  module:
    type: userpass
    userpass:
      username: user
      password: pass
`))
	if err != nil {
		t.Fatalf("DecodeAuthConfig() error = %v", err)
	}
	if got, want := config.AuthModules["module"].UserPass.Username, "user"; got != want {
		t.Fatalf("username = %q, want %q", got, want)
	}
}

func TestLoadAuthConfigWithHandler(t *testing.T) {
	ch, err := NewAuthConfigHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewAuthConfigHandler() error = %v", err)
	}

	if err := ch.ReloadAuthConfig("testdata/config-good.yaml", nil); err != nil {
		t.Errorf("error loading auth config: %s", err)
	}
}

func TestNewAuthConfigHandlerRequiresRegisterer(t *testing.T) {
	handler, err := NewAuthConfigHandler(nil)
	if err == nil {
		t.Fatal("NewAuthConfigHandler() error = nil, want error")
	}
	if handler != nil {
		t.Fatalf("NewAuthConfigHandler() handler = %v, want nil", handler)
	}
}

func TestLoadBadConfigs(t *testing.T) {
	ch, err := NewAuthConfigHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewAuthConfigHandler() error = %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "testdata/config-bad-auth-module.yaml",
			want:  "error parsing config file \"testdata/config-bad-auth-module.yaml\": yaml: unmarshal errors:\n  line 3: field pretendauth not found in type config.AuthModule",
		},
		{
			input: "testdata/config-bad-extra-field.yaml",
			want:  "error parsing config file \"testdata/config-bad-extra-field.yaml\": yaml: unmarshal errors:\n  line 8: field doesNotExist not found in type config.AuthModule",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := ch.ReloadAuthConfig(test.input, nil)
			if got == nil || got.Error() != test.want {
				t.Fatalf("ReloadAuthConfig(%q) = %v, want %s", test.input, got, test.want)
			}
		})
	}
}
