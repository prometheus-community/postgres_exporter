package metricmaps

import (
	"fmt"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type SemverSuite struct{}

var _ = Suite(&SemverSuite{})

func (s *SemverSuite) TestEncodeYAML(c *C) {
	sr, err := ParseSemverRange(">=1.0.0")
	c.Assert(err, IsNil)

	fmt.Println(sr)
	b, err := yaml.Marshal(&sr)
	c.Check(err, IsNil)
	fmt.Println(string(b))
}