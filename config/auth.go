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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
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

type AuthConfigHandler struct {
	sync.RWMutex
	AuthConfig *AuthConfig

	configReloadSuccess prometheus.Gauge
	configReloadSeconds prometheus.Gauge
}

func NewAuthConfigHandler(registerer prometheus.Registerer) (*AuthConfigHandler, error) {
	if registerer == nil {
		return nil, errors.New("registerer is required")
	}
	h := &AuthConfigHandler{
		AuthConfig: &AuthConfig{},
		configReloadSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "postgres_exporter",
			Name:      "config_last_reload_successful",
			Help:      "Postgres exporter config loaded successfully.",
		}),
		configReloadSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "postgres_exporter",
			Name:      "config_last_reload_success_timestamp_seconds",
			Help:      "Timestamp of the last successful configuration reload.",
		}),
	}
	registerer.MustRegister(h.configReloadSuccess, h.configReloadSeconds)

	return h, nil
}

func (ch *AuthConfigHandler) GetAuthConfig() *AuthConfig {
	ch.RLock()
	defer ch.RUnlock()
	return ch.AuthConfig
}

func (ch *AuthConfigHandler) ReloadAuthConfig(f string, logger *slog.Logger) error {
	var err error
	defer func() {
		ch.observeReload(err)
	}()

	config, err := LoadAuthConfig(f)
	if err != nil {
		return err
	}

	ch.SetAuthConfig(config)
	return nil
}

func (ch *AuthConfigHandler) observeReload(err error) {
	if ch.configReloadSuccess == nil {
		return
	}
	if err != nil {
		ch.configReloadSuccess.Set(0)
		return
	}
	ch.configReloadSuccess.Set(1)
	if ch.configReloadSeconds != nil {
		ch.configReloadSeconds.SetToCurrentTime()
	}
}

func LoadAuthConfig(f string) (*AuthConfig, error) {
	yamlReader, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("error opening config file %q: %s", f, err)
	}
	defer yamlReader.Close()

	config, err := DecodeAuthConfig(yamlReader)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %s", f, err)
	}
	return config, nil
}

func DecodeAuthConfig(r io.Reader) (*AuthConfig, error) {
	config := &AuthConfig{}
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (ch *AuthConfigHandler) SetAuthConfig(config *AuthConfig) {
	ch.Lock()
	ch.AuthConfig = config
	ch.Unlock()
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
