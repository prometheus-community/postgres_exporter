package builtin

import (
	"fmt"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MetricMapsSuite struct{}

var _ = Suite(&MetricMapsSuite{})

func (s *MetricMapsSuite) TestEncodeYAML(c *C) {
	data, err := yaml.Marshal(&builtin)
	c.Assert(err, IsNil)
	fmt.Println(string(data))
}