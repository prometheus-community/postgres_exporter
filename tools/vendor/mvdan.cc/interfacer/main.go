// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main // import "mvdan.cc/interfacer"

import (
	"flag"
	"fmt"
	"go/build"
	"os"

	"golang.org/x/tools/go/buildutil"

	"mvdan.cc/interfacer/check"
)

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags",
		buildutil.TagsFlagDoc)
}

func main() {
	flag.Parse()
	lines, err := check.CheckArgs(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, line := range lines {
		fmt.Println(line)
	}
}
