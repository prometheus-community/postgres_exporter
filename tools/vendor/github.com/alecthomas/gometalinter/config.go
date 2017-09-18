package main

import (
	"encoding/json"
	"runtime"
	"text/template"
	"time"
)

// Config for gometalinter. This can be loaded from a JSON file with --config.
type Config struct { // nolint: aligncheck
	// A map of linter name to "<command>:<pattern>".
	//
	// <command> should always include {path} as the target directory to execute. Globs in <command>
	// are expanded by gometalinter (not by the shell).
	Linters map[string]string

	// The set of linters that should be enabled.
	Enable  []string
	Disable []string

	// A map of linter name to message that is displayed. This is useful when linters display text
	// that is useful only in isolation, such as errcheck which just reports the construct.
	MessageOverride map[string]string
	Severity        map[string]string
	VendoredLinters bool
	Format          string
	Fast            bool
	Install         bool
	Update          bool
	Force           bool
	DownloadOnly    bool
	Debug           bool
	Concurrency     int
	Exclude         []string
	Include         []string
	Skip            []string
	Vendor          bool
	Cyclo           int
	LineLength      int
	MinConfidence   float64
	MinOccurrences  int
	MinConstLength  int
	DuplThreshold   int
	Sort            []string
	Test            bool
	Deadline        jsonDuration
	Errors          bool
	JSON            bool
	Checkstyle      bool
	EnableGC        bool
	Aggregate       bool
	EnableAll       bool
}

type jsonDuration time.Duration

func (td *jsonDuration) UnmarshalJSON(raw []byte) error {
	var durationAsString string
	if err := json.Unmarshal(raw, &durationAsString); err != nil {
		return err
	}
	duration, err := time.ParseDuration(durationAsString)
	*td = jsonDuration(duration)
	return err
}

// Duration returns the value as a time.Duration
func (td *jsonDuration) Duration() time.Duration {
	return time.Duration(*td)
}

// TODO: should be a field on Config struct
var formatTemplate = &template.Template{}

var sortKeys = []string{"none", "path", "line", "column", "severity", "message", "linter"}

// Configuration defaults.
var config = &Config{
	Format: "{{.Path}}:{{.Line}}:{{if .Col}}{{.Col}}{{end}}:{{.Severity}}: {{.Message}} ({{.Linter}})",

	Severity: map[string]string{
		"gotype":  "error",
		"test":    "error",
		"testify": "error",
		"vet":     "error",
	},
	MessageOverride: map[string]string{
		"errcheck":    "error return value not checked ({message})",
		"gocyclo":     "cyclomatic complexity {cyclo} of function {function}() is high (> {mincyclo})",
		"gofmt":       "file is not gofmted with -s",
		"goimports":   "file is not goimported",
		"safesql":     "potentially unsafe SQL statement",
		"structcheck": "unused struct field {message}",
		"unparam":     "parameter {message}",
		"varcheck":    "unused variable or constant {message}",
	},
	Enable:          defaultEnabled(),
	VendoredLinters: true,
	Concurrency:     runtime.NumCPU(),
	Cyclo:           10,
	LineLength:      80,
	MinConfidence:   0.8,
	MinOccurrences:  3,
	MinConstLength:  3,
	DuplThreshold:   50,
	Sort:            []string{"none"},
	Deadline:        jsonDuration(time.Second * 30),
}
