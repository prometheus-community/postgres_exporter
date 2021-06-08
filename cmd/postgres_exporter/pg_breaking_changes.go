package main

import (
	"github.com/blang/semver"
	"strings"
)

type BreakingChanges struct {
	Version string `yaml:"version"`
	ver     semver.Version
	Columns map[string]string `yaml:"columns"`
}

func (bc *BreakingChanges) ParseVerTolerant() error {
	bcVer, err := semver.ParseTolerant(bc.Version)
	if err != nil {
		return err
	}

	bc.ver = bcVer
	return nil
}

func (bc *BreakingChanges) FixColumns(query string) string {
	oldnew := make([]string, 0, 2*len(bc.Columns))
	for old := range bc.Columns {
		oldnew = append(oldnew, old, bc.Columns[old])
	}
	r := strings.NewReplacer(oldnew...)
	return r.Replace(query)
}
