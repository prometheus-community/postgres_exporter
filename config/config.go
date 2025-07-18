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
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v3"
)

var (
	configReloadSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "postgres_exporter",
		Name:      "config_last_reload_successful",
		Help:      "Postgres exporter config loaded successfully.",
	})

	configReloadSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "postgres_exporter",
		Name:      "config_last_reload_success_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
)

type Config struct {
	AuthModules map[string]AuthModule `yaml:"auth_modules"`
}

type AuthModule struct {
	Type     string   `yaml:"type"`
	UserPass UserPass `yaml:"userpass,omitempty"`
	// Add alternative auth modules here
	Options map[string]string `yaml:"options"`
}

type UserPass struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Handler struct {
	sync.RWMutex
	Config *Config
}

func (ch *Handler) GetConfig() *Config {
	ch.RLock()
	defer ch.RUnlock()
	return ch.Config
}

func (ch *Handler) ReloadConfig(f string, logger *slog.Logger) error {
	config := &Config{}
	var err error
	defer func() {
		if err != nil {
			configReloadSuccess.Set(0)
		} else {
			configReloadSuccess.Set(1)
			configReloadSeconds.SetToCurrentTime()
		}
	}()

	yamlReader, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("error opening config file %q: %s", f, err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(config); err != nil {
		return fmt.Errorf("error parsing config file %q: %s", f, err)
	}

	ch.Lock()
	ch.Config = config
	ch.Unlock()
	return nil
}

func (m AuthModule) ConfigureTarget(target string) (DSN, error) {
	dsn, err := dsnFromString(target)
	if err != nil {
		return DSN{}, err
	}

	// Set the credentials from the authentication module
	// TODO(@sysadmind): What should the order of precedence be?
	if m.Type == "userpass" {
		if m.UserPass.Username != "" {
			dsn.username = m.UserPass.Username
		}
		if m.UserPass.Password != "" {
			dsn.password = m.UserPass.Password
		}
	}

	for k, v := range m.Options {
		dsn.query.Set(k, v)
	}

	return dsn, nil
}
