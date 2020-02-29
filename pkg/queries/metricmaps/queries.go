package metricmaps

import (
	"fmt"
	"github.com/blang/semver"
)

// ColumnUsage should be one of several enum values which describe how a
// queried row is to be converted to a Prometheus metric.
type ColumnUsage string

const (
	DISCARD      ColumnUsage = "DISCARD" // Ignore this column
	LABEL        ColumnUsage = "LABEL" // Use this column as a label
	COUNTER      ColumnUsage = "COUNTER" // Use this column as a counter
	GAUGE        ColumnUsage = "GAUGE" // Use this column as a gauge
	MAPPEDMETRIC ColumnUsage = "MAPPEDMETRIC" // Use this column with the supplied mapping of text values
	DURATION     ColumnUsage = "DURATION" // This column should be interpreted as a text duration (and converted to milliseconds)
)

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (cu *ColumnUsage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}

	var columnUsage ColumnUsage
	switch value {
	case "DISCARD",
		 "LABEL",
		 "COUNTER",
		 "GAUGE",
		 "MAPPEDMETRIC",
		 "DURATION":
		columnUsage = ColumnUsage(value)
	default:
		return fmt.Errorf("value is not a valid ColumnUsage value: %s", value)
	}

	*cu = columnUsage
	return nil
}

// ColumnMapping is the user-friendly representation of a prometheus descriptor map
type ColumnMapping struct {
	Usage             ColumnUsage        `yaml:"usage"`
	Description       string             `yaml:"description"`
	Mapping           map[string]float64 `yaml:"metric_mapping"` // Optional column mapping for MAPPEDMETRIC
	SupportedVersions *SemverRange       `yaml:"pg_version"`     // Semantic version ranges which are supported. Unsupported columns are not queried (internally converted to DISCARD).
}

// UnmarshalYAML implements yaml.Unmarshaller
func (cm *ColumnMapping) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ColumnMapping
	return unmarshal((*plain)(cm))
}

type Mapping map[string]MappingOptions

type UserQuery struct {
	Query        string    `yaml:"query"`
	Metrics      []Mapping `yaml:"metrics"`
	Master       bool      `yaml:"master"`        // Querying only for master database
	CacheSeconds uint64    `yaml:"cache_seconds"` // Number of seconds to cache the ExporterNamespaceLabel result metrics for.
}

type UserQueries map[string]UserQuery

// OverrideQuery are run in-place of simple ExporterNamespaceLabel look ups, and provide
// advanced functionality. But they have a tendency to postgres version specific.
// There aren't too many versions, so we simply store customized versions using
// the semver matching we do for columns.
type OverrideQuery struct {
	VersionRange semver.Range
	Query        string
}

// SemverRange implements YAML marshalling for semver.Range
type SemverRange struct {
	r string
	semver.Range
}

// MustParseSemverRange parses a semver
func MustParseSemverRange(s string) *SemverRange {
	r, err := ParseSemverRange(s)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseSemverRange parses a semver
func ParseSemverRange(s string) (*SemverRange, error) {
	r, err := semver.ParseRange(s)
	if err != nil {
		return nil, err
	}
	return &SemverRange{s, r}, nil
}

func (sr *SemverRange) String() string {
	if sr != nil {
		return sr.r
	} else {
		return "(any)"
	}
}

// MarshalYAML implements yaml.Marshaller
func (sr *SemverRange) MarshalYAML() (interface{}, error) {
	if sr == nil {
		return nil, nil
	}
	return sr.r, nil
}

// UnmarshalYAML implements yaml.Unmarshaller
func (sr *SemverRange) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	var rangeStr string
	if err = unmarshal(&rangeStr); err != nil {
		return err
	}

	sr, err = ParseSemverRange(rangeStr)
	if err != nil {
		return err
	}

	return nil
}