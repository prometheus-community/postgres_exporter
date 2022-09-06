// postgres_exporter

//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/prometheus/promu"
	_ "github.com/reviewdog/reviewdog/cmd/reviewdog"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
