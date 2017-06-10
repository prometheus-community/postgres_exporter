package main

import (
	"runtime"
	"text/template"
	"time"

	"gopkg.in/alecthomas/kingpin.v3-unstable"
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
	Deadline        time.Duration `json:"-"`
	Errors          bool
	JSON            bool
	Checkstyle      bool
	EnableGC        bool
	Aggregate       bool

	DeadlineJSONCrutch string `json:"Deadline"`
}

// Configuration defaults.
var (
	vetRe = `^(?:vet:.*?\.go:\s+(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*))|(?:(?P<path>.*?\.go):(?P<line>\d+):\s*(?P<message>.*))$`

	predefinedPatterns = map[string]string{
		"PATH:LINE:COL:MESSAGE": `^(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$`,
		"PATH:LINE:MESSAGE":     `^(?P<path>.*?\.go):(?P<line>\d+):\s*(?P<message>.*)$`,
	}
	formatTemplate = &template.Template{}
	installMap     = map[string]string{
		"aligncheck":  "github.com/opennota/check/cmd/aligncheck",
		"deadcode":    "github.com/tsenart/deadcode",
		"dupl":        "github.com/mibk/dupl",
		"errcheck":    "github.com/kisielk/errcheck",
		"gas":         "github.com/GoASTScanner/gas",
		"goconst":     "github.com/jgautheron/goconst/cmd/goconst",
		"gocyclo":     "github.com/alecthomas/gocyclo",
		"goimports":   "golang.org/x/tools/cmd/goimports",
		"golint":      "github.com/golang/lint/golint",
		"gosimple":    "honnef.co/go/tools/cmd/gosimple",
		"gotype":      "golang.org/x/tools/cmd/gotype",
		"ineffassign": "github.com/gordonklaus/ineffassign",
		"interfacer":  "github.com/mvdan/interfacer/cmd/interfacer",
		"lll":         "github.com/walle/lll/cmd/lll",
		"misspell":    "github.com/client9/misspell/cmd/misspell",
		"safesql":     "github.com/stripe/safesql",
		"staticcheck": "honnef.co/go/tools/cmd/staticcheck",
		"structcheck": "github.com/opennota/check/cmd/structcheck",
		"unconvert":   "github.com/mdempsky/unconvert",
		"unparam":     "github.com/mvdan/unparam",
		"unused":      "honnef.co/go/tools/cmd/unused",
		"varcheck":    "github.com/opennota/check/cmd/varcheck",
	}
	acceptsEllipsis = map[string]bool{
		"aligncheck":  true,
		"errcheck":    true,
		"golint":      true,
		"gosimple":    true,
		"interfacer":  true,
		"staticcheck": true,
		"structcheck": true,
		"test":        true,
		"varcheck":    true,
		"unconvert":   true,
	}
	slowLinters = []string{"structcheck", "varcheck", "errcheck", "aligncheck", "testify", "test", "interfacer", "unconvert", "deadcode", "safesql", "staticcheck", "unparam", "unused", "gosimple"}
	sortKeys    = []string{"none", "path", "line", "column", "severity", "message", "linter"}

	// Linter definitions.
	linterDefinitions = map[string]string{
		"aligncheck":  `aligncheck {path}:^(?:[^:]+: )?(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.+)$`,
		"deadcode":    `deadcode {path}:^deadcode: (?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$`,
		"dupl":        `dupl -plumbing -threshold {duplthreshold} {path}/*.go:^(?P<path>.*?\.go):(?P<line>\d+)-\d+:\s*(?P<message>.*)$`,
		"errcheck":    `errcheck -abspath {path}:PATH:LINE:COL:MESSAGE`,
		"gas":         `gas -fmt=csv {path}/*.go:^(?P<path>.*?\.go),(?P<line>\d+),(?P<message>[^,]+,[^,]+,[^,]+)`,
		"goconst":     `goconst -min-occurrences {min_occurrences} -min-length {min_const_length} {path}:PATH:LINE:COL:MESSAGE`,
		"gocyclo":     `gocyclo -over {mincyclo} {path}:^(?P<cyclo>\d+)\s+\S+\s(?P<function>\S+)\s+(?P<path>.*?\.go):(?P<line>\d+):(\d+)$`,
		"gofmt":       `gofmt -l -s {path}/*.go:^(?P<path>.*?\.go)$`,
		"goimports":   `goimports -l {path}/*.go:^(?P<path>.*?\.go)$`,
		"golint":      "golint -min_confidence {min_confidence} {path}:PATH:LINE:COL:MESSAGE",
		"gosimple":    "gosimple {path}:PATH:LINE:COL:MESSAGE",
		"gotype":      "gotype -e {tests=-a} {path}:PATH:LINE:COL:MESSAGE",
		"ineffassign": `ineffassign -n {path}:PATH:LINE:COL:MESSAGE`,
		"interfacer":  `interfacer {path}:PATH:LINE:COL:MESSAGE`,
		"lll":         `lll -g -l {maxlinelength} {path}/*.go:PATH:LINE:MESSAGE`,
		"misspell":    "misspell -j 1 {path}/*.go:PATH:LINE:COL:MESSAGE",
		"safesql":     `safesql {path}:^- (?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+)$`,
		"staticcheck": "staticcheck {path}:PATH:LINE:COL:MESSAGE",
		"structcheck": `structcheck {tests=-t} {path}:^(?:[^:]+: )?(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.+)$`,
		"test":        `go test {path}:^--- FAIL: .*$\s+(?P<path>.*?\.go):(?P<line>\d+): (?P<message>.*)$`,
		"testify":     `go test {path}:Location:\s+(?P<path>.*?\.go):(?P<line>\d+)$\s+Error:\s+(?P<message>[^\n]+)`,
		"unconvert":   "unconvert {path}:PATH:LINE:COL:MESSAGE",
		"unparam":     `unparam {path}:PATH:LINE:COL:MESSAGE`,
		"unused":      `unused {path}:PATH:LINE:COL:MESSAGE`,
		"varcheck":    `varcheck {path}:^(?:[^:]+: )?(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$`,
		"vet":         `go tool vet {path}/*.go:` + vetRe,
		"vetshadow":   `go tool vet --shadow {path}/*.go:` + vetRe,
	}

	pathsArg = kingpin.Arg("path", "Directories to lint. Defaults to \".\". <path>/... will recurse.").Strings()

	config = &Config{
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
		Enable: []string{
			"aligncheck",
			"deadcode",
			"errcheck",
			"gas",
			"goconst",
			"gocyclo",
			"golint",
			"gosimple",
			"gotype",
			"ineffassign",
			"interfacer",
			"staticcheck",
			"structcheck",
			"unconvert",
			"varcheck",
			"vet",
			"vetshadow",
		},
		VendoredLinters: true,
		Concurrency:     runtime.NumCPU(),
		Cyclo:           10,
		LineLength:      80,
		MinConfidence:   0.8,
		MinOccurrences:  3,
		MinConstLength:  3,
		DuplThreshold:   50,
		Sort:            []string{"none"},
		Deadline:        time.Second * 30,
	}
)
