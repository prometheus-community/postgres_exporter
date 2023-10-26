//go:build manual
// +build manual

package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//go:embed percona-reference-metrics.txt
var referenceMetrics string

// TestReferenceCompatibility checks that exposed metrics are not missed.
//
// Used to make sure that metrics are present after updating from upstream.
// You need you run exporter locally on port 42002.
func TestReferenceCompatibility(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", "http://localhost:42000/metrics", nil)
	assert.Nil(t, err)
	req.SetBasicAuth("pmm", "/agent_id/825dcdbf-af1c-4eb4-9e96-21699aa6ff7b")
	resp, err := client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	currentMetricsBytes, err := os.ReadAll(resp.Body)
	assert.Nil(t, err)

	currentMetrics := toMap(t, string(currentMetricsBytes))
	referenceMetrics := toMap(t, referenceMetrics)

	//remove matches
	for m := range currentMetrics {
		_, found := referenceMetrics[m]
		if found {
			delete(referenceMetrics, m)
			delete(currentMetrics, m)
		}
	}

	fmt.Printf("Extra metrics [%d]:\n", len(currentMetrics))
	for _, metric := range sortedKeys(currentMetrics) {
		fmt.Printf("\t%s\n", metric)
	}
	if len(referenceMetrics) != 0 {
		fmt.Printf("Not Supported metrics [%d]:\n", len(referenceMetrics))
		for _, metric := range sortedKeys(referenceMetrics) {
			fmt.Printf("\t%s\n", metric)
		}
		assert.FailNowf(t, "Found not supported metrics", "Count: %d", len(referenceMetrics))
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func toMap(t *testing.T, rawMetrics string) map[string]string {
	result := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(rawMetrics))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		next := scanner.Text()
		isComment := strings.HasPrefix(next, "#")
		if isComment {
			continue
		}
		next = cleanKeyOrValue(next)
		if next != "" {
			items := strings.Split(next, " ")
			if len(items) > 1 {
				result[items[0]] = items[1]
			} else {
				fmt.Println("WARN: ")
			}
		}
	}

	return result
}

func cleanKeyOrValue(s string) (res string) {
	res = s

	itemsToIgnore := []string{
		"example-queries",
	}

	for _, each := range itemsToIgnore {
		if strings.Contains(s, each) {
			return ""
		}
	}

	regexpsToRemove := []*regexp.Regexp{
		regexp.MustCompile(`[+-]?(\d*[.])?\d+(e[+-]?\d*)?`),
		regexp.MustCompile(`\d*\.\d*\.\d*\.\d*:\d*`),
		regexp.MustCompile(`go1.\d*.\d*`),
		regexp.MustCompile(`filename=".*",`),
		regexp.MustCompile(`hashsum=".*"`),
	}
	for _, each := range regexpsToRemove {
		res = each.ReplaceAllString(res, "")
	}

	stringsToRemove := []string{
		"PostgreSQL 11.15 (Debian 11.15-1.pgdg90+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
		"PostgreSQL 11.16 (Debian 11.16-1.pgdg90+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
		"collector=\"exporter\",",
		"fastpath function call",
		"idle in transaction (aborted)",
		"idle in transaction",
		"+Inf",
		"0.0.1",
		"collector=\"custom_query.mr\",",
		"datname=\"pmm-managed\"",
		"datname=\"pmm-agent\"",
	}
	for _, each := range stringsToRemove {
		res = strings.ReplaceAll(res, each, "")
	}

	return
}
