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

// +build !integration

package main

import (
	"testing"

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
		d: `Desc{fqName: "pg_settings_seconds_fixture_metric_seconds", help: "Foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_milliseconds_fixture_metric_seconds", help: "Foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_eight_kb_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_16_kb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_16_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_32_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_64_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_bool_on_fixture_metric", help: "Foo foo foo", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_bool_off_fixture_metric", help: "Foo foo foo", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_special_minus_one_value_seconds", help: "foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
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
		d: `Desc{fqName: "pg_settings_rds_rds_superuser_reserved_connections", help: "Sets the number of connection slots reserved for rds_superusers.", constLabels: {}, variableLabels: []}`,
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
			err:  `Unknown unit for runtime variable: "nonexistent"`,
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
			if r.(error).Error() != `Unknown unit for runtime variable: "nonexistent"` {
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

func Test_pgSetting_normaliseUnit(t *testing.T) {
	type fields struct {
		name      string
		setting   string
		unit      string
		shortDesc string
		vartype   string
	}
	tests := []struct {
		name     string
		fields   fields
		wantVal  float64
		wantUnit string
		wantErr  bool
	}{
		{
			name: "Seconds",
			fields: fields{
				name:      "seconds_fixture_metric",
				setting:   "5",
				unit:      "s",
				shortDesc: "Foo foo foo",
				vartype:   "integer",
			},
			wantVal:  5,
			wantUnit: "seconds",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_seconds_fixture_metric_seconds", help: "Foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
			// v: 5,
		},
		{
			name: "Milliseconds",
			fields: fields{
				name:      "milliseconds_fixture_metric",
				setting:   "5000",
				unit:      "ms",
				shortDesc: "Foo foo foo",
				vartype:   "integer",
			},
			wantVal:  5,
			wantUnit: "seconds",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_milliseconds_fixture_metric_seconds", help: "Foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
			// v: 5,
		},
		{
			name: "8KB",
			fields: fields{
				name:      "eight_kb_fixture_metric",
				setting:   "17",
				unit:      "8kB",
				shortDesc: "Foo foo foo",
				vartype:   "integer",
			},
			wantVal:  139264,
			wantUnit: "bytes",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_eight_kb_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
			// v: 139264,
		},

		{
			name: "16KB real",
			fields: fields{
				name:      "16_kb_real_fixture_metric",
				setting:   "3.0",
				unit:      "16kB",
				shortDesc: "Foo foo foo",
				vartype:   "real",
			},
			wantVal:  49152,
			wantUnit: "bytes",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_16_kb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
			// v: 49152,
		},
		{
			name: "16MB real",
			fields: fields{
				name:      "16_mb_real_fixture_metric",
				setting:   "3.0",
				unit:      "16MB",
				shortDesc: "Foo foo foo",
				vartype:   "real",
			},
			wantVal:  5.0331648e+07,
			wantUnit: "bytes",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_16_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
			// v: 5.0331648e+07,
		},
		{
			name: "32MB real",
			fields: fields{
				name:      "32_mb_real_fixture_metric",
				setting:   "3.0",
				unit:      "32MB",
				shortDesc: "Foo foo foo",
				vartype:   "real",
			},
			wantVal:  1.00663296e+08,
			wantUnit: "bytes",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_32_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
			// v: 1.00663296e+08,
		},
		{
			name: "64MB real",
			fields: fields{
				name:      "64_mb_real_fixture_metric",
				setting:   "3.0",
				unit:      "64MB",
				shortDesc: "Foo foo foo",
				vartype:   "real",
			},
			wantVal:  2.01326592e+08,
			wantUnit: "bytes",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_64_mb_real_fixture_metric_bytes", help: "Foo foo foo [Units converted to bytes.]", constLabels: {}, variableLabels: []}`,
			// v: 2.01326592e+08,
		},

		// TODO: normalize doesn't handle boolean, that is currently handled in pgSetting.metric
		// {
		// 	name: "Boolean on",
		// 	fields: fields{
		// 		name:      "bool_on_fixture_metric",
		// 		setting:   "on",
		// 		unit:      "",
		// 		shortDesc: "Foo foo foo",
		// 		vartype:   "bool",
		// 	},
		// 	wantVal:  1,
		// 	wantUnit: "",
		// 	wantErr:  false,
		// 	// d: `Desc{fqName: "pg_settings_bool_on_fixture_metric", help: "Foo foo foo", constLabels: {}, variableLabels: []}`,
		// 	// v: 1,
		// },

		// {
		// 	name: "Boolean off",
		// 	fields: fields{
		// 		name:      "bool_off_fixture_metric",
		// 		setting:   "off",
		// 		unit:      "",
		// 		shortDesc: "Foo foo foo",
		// 		vartype:   "bool",
		// 	},
		// 	wantVal:  0,
		// 	wantUnit: "",
		// 	wantErr:  false,
		// 	// d: `Desc{fqName: "pg_settings_bool_off_fixture_metric", help: "Foo foo foo", constLabels: {}, variableLabels: []}`,
		// 	// v: 0,
		// },
		{
			name: "Special -1",
			fields: fields{
				name:      "special_minus_one_value",
				setting:   "-1",
				unit:      "d",
				shortDesc: "Foo foo foo",
				vartype:   "integer",
			},
			wantVal:  -1,
			wantUnit: "seconds",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_special_minus_one_value_seconds", help: "foo foo foo [Units converted to seconds.]", constLabels: {}, variableLabels: []}`,
			// v: -1,
		},
		{
			name: "RDS superuser",
			fields: fields{
				name:      "rds.rds_superuser_reserved_connections",
				setting:   "2",
				unit:      "",
				shortDesc: "Sets the number of connection slots reserved for rds_superusers.",
				vartype:   "integer",
			},
			wantVal:  2,
			wantUnit: "",
			wantErr:  false,
			// d: `Desc{fqName: "pg_settings_rds_rds_superuser_reserved_connections", help: "Sets the number of connection slots reserved for rds_superusers.", constLabels: {}, variableLabels: []}`,
			// v: 2,
		},
		{
			name: "Unknown unit",
			fields: fields{
				name:      "10",
				setting:   "nonexistent",
				unit:      "",
				shortDesc: "Foo foo foo",
				vartype:   "integer",
			},
			wantVal:  0,
			wantUnit: "",
			wantErr:  true,
			// err:  `Unknown unit for runtime variable: "nonexistent"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &pgSetting{
				name:      tt.fields.name,
				setting:   tt.fields.setting,
				unit:      tt.fields.unit,
				shortDesc: tt.fields.shortDesc,
				vartype:   tt.fields.vartype,
			}
			gotVal, gotUnit, err := s.normaliseUnit()
			if (err != nil) != tt.wantErr {
				t.Errorf("pgSetting.normaliseUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("pgSetting.normaliseUnit() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotUnit != tt.wantUnit {
				t.Errorf("pgSetting.normaliseUnit() gotUnit = %v, want %v", gotUnit, tt.wantUnit)
			}
		})
	}
}
