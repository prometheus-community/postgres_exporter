// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main // import "mvdan.cc/unparam"

import (
	"flag"
	"fmt"
	"go/build"
	"os"

	"golang.org/x/tools/go/buildutil"

	"mvdan.cc/unparam/check"
)

var (
	algo     = flag.String("algo", "cha", `call graph construction algorithm (cha, rta).
in general, use cha for libraries, and rta for programs with main packages.`)
	tests    = flag.Bool("tests", true, "include tests")
	exported = flag.Bool("exported", false, "inspect exported functions")
	debug    = flag.Bool("debug", false, "debug prints")
)

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags",
		buildutil.TagsFlagDoc)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: unparam [flags] [package ...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	warns, err := check.UnusedParams(*tests, *algo, *exported, *debug, flag.Args()...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, warn := range warns {
		fmt.Println(warn)
	}
}
