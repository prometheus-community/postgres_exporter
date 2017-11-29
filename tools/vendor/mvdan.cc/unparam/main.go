// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main // import "mvdan.cc/unparam"

import (
	"flag"
	"fmt"
	"os"

	"mvdan.cc/unparam/check"
)

var (
	tests    = flag.Bool("tests", true, "include tests")
	exported = flag.Bool("exported", false, "inspect exported functions")
	debug    = flag.Bool("debug", false, "debug prints")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: unparam [flags] [package ...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	warns, err := check.UnusedParams(*tests, *exported, *debug, flag.Args()...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, warn := range warns {
		fmt.Println(warn)
	}
}
