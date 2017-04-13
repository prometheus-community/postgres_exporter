// +build !integration

package main

import (
	dto "github.com/prometheus/client_model/go"
	. "gopkg.in/check.v1"
)

type PgSettingSuite struct{}

var _ = Suite(&PgSettingSuite{})

var fixtures = []fixture{
	fixture{
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
		d: "Desc{fqName: \"pg_settings_seconds_fixture_metric_seconds\", help: \"Foo foo foo [Units converted to seconds.]\", constLabels: {}, variableLabels: []}",
		v: 5,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_milliseconds_fixture_metric_seconds\", help: \"Foo foo foo [Units converted to seconds.]\", constLabels: {}, variableLabels: []}",
		v: 5,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_eight_kb_fixture_metric_bytes\", help: \"Foo foo foo [Units converted to bytes.]\", constLabels: {}, variableLabels: []}",
		v: 139264,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_16_mb_real_fixture_metric_bytes\", help: \"Foo foo foo [Units converted to bytes.]\", constLabels: {}, variableLabels: []}",
		v: 5.0331648e+07,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_bool_on_fixture_metric\", help: \"Foo foo foo\", constLabels: {}, variableLabels: []}",
		v: 1,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_bool_off_fixture_metric\", help: \"Foo foo foo\", constLabels: {}, variableLabels: []}",
		v: 0,
	},
	fixture{
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
		d: "Desc{fqName: \"pg_settings_special_minus_one_value_seconds\", help: \"foo foo foo [Units converted to seconds.]\", constLabels: {}, variableLabels: []}",
		v: -1,
	},
	fixture{
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
		m := f.p.metric()
		m.Write(d)

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
