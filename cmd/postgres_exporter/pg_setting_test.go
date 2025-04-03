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

//go:build !integration
// +build !integration

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	. "gopkg.in/check.v1"
)

type PgSettingSuite struct{}

var _ = Suite(&PgSettingSuite{})

var fixtures = []fixture{
	{
		p: pgSetting{
			name:      "seconds_fixture_metric",
			setting:   "5",
			unit:      "s",
			shortDesc: "Foo foo foo",
			vartype:   "integer",
		},
		n: normalised{
			val:  5,
			unit: "seconds",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_seconds_fixture_metric_seconds", help: "Server Parameter: seconds_fixture_metric [Units converted to seconds.]", constLabels: {}, variableLabels: {}}`,
		v: 5,
	},
	{
		p: pgSetting{
			name:      "milliseconds_fixture_metric",
			setting:   "5000",
			unit:      "ms",
			shortDesc: "Foo foo foo",
			vartype:   "integer",
		},
		n: normalised{
			val:  5,
			unit: "seconds",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_milliseconds_fixture_metric_seconds", help: "Server Parameter: milliseconds_fixture_metric [Units converted to seconds.]", constLabels: {}, variableLabels: {}}`,
		v: 5,
	},
	{
		p: pgSetting{
			name:      "eight_kb_fixture_metric",
			setting:   "17",
			unit:      "8kB",
			shortDesc: "Foo foo foo",
			vartype:   "integer",
		},
		n: normalised{
			val:  139264,
			unit: "bytes",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_eight_kb_fixture_metric_bytes", help: "Server Parameter: eight_kb_fixture_metric [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
		v: 139264,
	},
	{
		p: pgSetting{
			name:      "16_kb_real_fixture_metric",
			setting:   "3.0",
			unit:      "16kB",
			shortDesc: "Foo foo foo",
			vartype:   "real",
		},
		n: normalised{
			val:  49152,
			unit: "bytes",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_16_kb_real_fixture_metric_bytes", help: "Server Parameter: 16_kb_real_fixture_metric [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
		v: 49152,
	},
	{
		p: pgSetting{
			name:      "16_mb_real_fixture_metric",
			setting:   "3.0",
			unit:      "16MB",
			shortDesc: "Foo foo foo",
			vartype:   "real",
		},
		n: normalised{
			val:  5.0331648e+07,
			unit: "bytes",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_16_mb_real_fixture_metric_bytes", help: "Server Parameter: 16_mb_real_fixture_metric [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
		v: 5.0331648e+07,
	},
	{
		p: pgSetting{
			name:      "32_mb_real_fixture_metric",
			setting:   "3.0",
			unit:      "32MB",
			shortDesc: "Foo foo foo",
			vartype:   "real",
		},
		n: normalised{
			val:  1.00663296e+08,
			unit: "bytes",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_32_mb_real_fixture_metric_bytes", help: "Server Parameter: 32_mb_real_fixture_metric [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
		v: 1.00663296e+08,
	},
	{
		p: pgSetting{
			name:      "64_mb_real_fixture_metric",
			setting:   "3.0",
			unit:      "64MB",
			shortDesc: "Foo foo foo",
			vartype:   "real",
		},
		n: normalised{
			val:  2.01326592e+08,
			unit: "bytes",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_64_mb_real_fixture_metric_bytes", help: "Server Parameter: 64_mb_real_fixture_metric [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
		v: 2.01326592e+08,
	},
	{
		p: pgSetting{
			name:      "bool_on_fixture_metric",
			setting:   "on",
			unit:      "",
			shortDesc: "Foo foo foo",
			vartype:   "bool",
		},
		n: normalised{
			val:  1,
			unit: "",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_bool_on_fixture_metric", help: "Server Parameter: bool_on_fixture_metric", constLabels: {}, variableLabels: {}}`,
		v: 1,
	},
	{
		p: pgSetting{
			name:      "bool_off_fixture_metric",
			setting:   "off",
			unit:      "",
			shortDesc: "Foo foo foo",
			vartype:   "bool",
		},
		n: normalised{
			val:  0,
			unit: "",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_bool_off_fixture_metric", help: "Server Parameter: bool_off_fixture_metric", constLabels: {}, variableLabels: {}}`,
		v: 0,
	},
	{
		p: pgSetting{
			name:      "special_minus_one_value",
			setting:   "-1",
			unit:      "d",
			shortDesc: "foo foo foo",
			vartype:   "integer",
		},
		n: normalised{
			val:  -1,
			unit: "seconds",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_special_minus_one_value_seconds", help: "Server Parameter: special_minus_one_value [Units converted to seconds.]", constLabels: {}, variableLabels: {}}`,
		v: -1,
	},
	{
		p: pgSetting{
			name:      "rds.rds_superuser_reserved_connections",
			setting:   "2",
			unit:      "",
			shortDesc: "Sets the number of connection slots reserved for rds_superusers.",
			vartype:   "integer",
		},
		n: normalised{
			val:  2,
			unit: "",
			err:  "",
		},
		d: `Desc{fqName: "pg_settings_rds_rds_superuser_reserved_connections", help: "Server Parameter: rds.rds_superuser_reserved_connections", constLabels: {}, variableLabels: {}}`,
		v: 2,
	},
	{
		p: pgSetting{
			name:      "unknown_unit",
			setting:   "10",
			unit:      "nonexistent",
			shortDesc: "foo foo foo",
			vartype:   "integer",
		},
		n: normalised{
			val:  10,
			unit: "",
			err:  `unknown unit for runtime variable: "nonexistent"`,
		},
	},
}

func (s *PgSettingSuite) TestNormaliseUnit(c *C) {
	for _, f := range fixtures {
		switch f.p.vartype {
		case "integer", "real":
			val, unit, err := f.p.normaliseUnit()

			c.Check(val, Equals, f.n.val)
			c.Check(unit, Equals, f.n.unit)

			if err == nil {
				c.Check("", Equals, f.n.err)
			} else {
				c.Check(err.Error(), Equals, f.n.err)
			}
		}
	}
}

func (s *PgSettingSuite) TestMetric(c *C) {
	defer func() {
		if r := recover(); r != nil {
			if r.(error).Error() != `unknown unit for runtime variable: "nonexistent"` {
				panic(r)
			}
		}
	}()

	for _, f := range fixtures {
		d := &dto.Metric{}
		m := f.p.metric(prometheus.Labels{})
		m.Write(d) // nolint: errcheck

		c.Check(m.Desc().String(), Equals, f.d)
		c.Check(d.GetGauge().GetValue(), Equals, f.v)
	}
}

type normalised struct {
	val  float64
	unit string
	err  string
}

type fixture struct {
	p pgSetting
	n normalised
	d string
	v float64
}
