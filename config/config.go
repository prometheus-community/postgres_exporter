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
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

var (
	configReloadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "postgres_exporter",
		Name:      "config_last_reload_successful",
		Help:      "Postgres exporter config loaded successfully.",
	})

	configReloadSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "postgres_exporter",
		Name:      "config_last_reload_success_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
)

func init() {
	prometheus.MustRegister(configReloadSuccess)
	prometheus.MustRegister(configReloadSeconds)
}

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

type ConfigHandler struct {
	sync.RWMutex
	Config *Config
}

func (ch *ConfigHandler) GetConfig() *Config {
	ch.RLock()
	defer ch.RUnlock()
	return ch.Config
}

func (ch *ConfigHandler) ReloadConfig(f string, logger log.Logger) error {
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
		return fmt.Errorf("Error opening config file %q: %s", f, err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(config); err != nil {
		return fmt.Errorf("Error parsing config file %q: %s", f, err)
	}

	ch.Lock()
	ch.Config = config
	ch.Unlock()
	return nil
}

func (m AuthModule) ConfigureTarget(target string) (string, error) {
	// ip:port urls do not parse properly and that is the typical way users interact with postgres
	t := fmt.Sprintf("exporter://%s", target)
	u, err := url.Parse(t)
	if err != nil {
		return "", err
	}

	if m.Type == "userpass" {
		u.User = url.UserPassword(m.UserPass.Username, m.UserPass.Password)
	}

	query := u.Query()
	for k, v := range m.Options {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	parsed := u.String()
	trim := strings.TrimPrefix(parsed, "exporter://")

	return trim, nil
}
