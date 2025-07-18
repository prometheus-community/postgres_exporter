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
	"testing"
)

func TestLoadConfig(t *testing.T) {
	ch := &Handler{
		Config: &Config{},
	}

	err := ch.ReloadConfig("testdata/config-good.yaml", nil)
	if err != nil {
		t.Errorf("error loading config: %s", err)
	}
}

func TestLoadBadConfigs(t *testing.T) {
	ch := &Handler{
		Config: &Config{},
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
			got := ch.ReloadConfig(test.input, nil)
			if got == nil || got.Error() != test.want {
				t.Fatalf("ReloadConfig(%q) = %v, want %s", test.input, got, test.want)
			}
		})
	}
}
