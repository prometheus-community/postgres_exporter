// Copyright 2021 The Prometheus Authors
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
	"strings"

	"github.com/blang/semver"
)

type BreakingChanges struct {
	Version string `yaml:"version"`
	ver     semver.Version
	Columns map[string]string `yaml:"columns"`
}

func (bc *BreakingChanges) ParseVerTolerant() error {
	bcVer, err := semver.ParseTolerant(bc.Version)
	if err != nil {
		return err
	}

	bc.ver = bcVer
	return nil
}

func (bc *BreakingChanges) FixColumns(query string) string {
	// nolint: golint
	// 2 because old - new
	oldnew := make([]string, 0, 2*len(bc.Columns))
	for old := range bc.Columns {
		oldnew = append(oldnew, old, bc.Columns[old])
	}
	r := strings.NewReplacer(oldnew...)
	return r.Replace(query)
}
