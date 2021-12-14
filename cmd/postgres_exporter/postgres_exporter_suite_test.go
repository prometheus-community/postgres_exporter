package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPostgresExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgresExporter Suite")
}
