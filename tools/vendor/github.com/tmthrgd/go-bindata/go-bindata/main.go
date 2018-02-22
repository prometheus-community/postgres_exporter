// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tmthrgd/go-bindata"
	"github.com/tmthrgd/go-bindata/internal/identifier"
)

func must(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "go-bindata: %v\n", err)
	os.Exit(1)
}

func main() {
	genOpts, findOpts, output := parseArgs()

	var all bindata.Files

	for i := 0; i < flag.NArg(); i++ {
		var path string
		path, findOpts.Recursive = parseInput(flag.Arg(i))

		files, err := bindata.FindFiles(path, findOpts)
		must(err)

		all = append(all, files...)
	}

	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	must(err)

	defer f.Close()

	must(all.Generate(f, genOpts))
}

// parseArgs create s a new, filled configuration instance
// by reading and parsing command line options.
//
// This function exits the program with an error, if
// any of the command line options are incorrect.
func parseArgs() (genOpts *bindata.GenerateOptions, findOpts *bindata.FindFilesOptions, output string) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	var version bool
	flag.BoolVar(&version, "version", false, "Displays version information.")

	flag.StringVar(&output, "o", "./bindata.go", "Optional name of the output file to be generated.")

	genOpts = &bindata.GenerateOptions{
		Package:        "main",
		MemCopy:        true,
		Compress:       true,
		Metadata:       true,
		Restore:        true,
		AssetDir:       true,
		DecompressOnce: true,
	}
	findOpts = new(bindata.FindFilesOptions)

	var noMemCopy, noCompress, noMetadata bool
	var mode uint
	flag.BoolVar(&genOpts.Debug, "debug", genOpts.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.BoolVar(&genOpts.Dev, "dev", genOpts.Dev, "Similar to debug, but does not emit absolute paths. Expects a rootDir variable to already exist in the generated code's package.")
	flag.StringVar(&genOpts.Tags, "tags", genOpts.Tags, "Optional set of build tags to include.")
	flag.StringVar(&findOpts.Prefix, "prefix", "", "Optional path prefix to strip off asset names.")
	flag.StringVar(&genOpts.Package, "pkg", genOpts.Package, "Package name to use in the generated code.")
	flag.BoolVar(&noMemCopy, "nomemcopy", !genOpts.MemCopy, "Use a .rodata hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
	flag.BoolVar(&noCompress, "nocompress", !genOpts.Compress, "Assets will *not* be GZIP compressed when this flag is specified.")
	flag.BoolVar(&noMetadata, "nometadata", !genOpts.Metadata, "Assets will not preserve size, mode, and modtime info.")
	flag.UintVar(&mode, "mode", uint(genOpts.Mode), "Optional file mode override for all files.")
	flag.Int64Var(&genOpts.ModTime, "modtime", genOpts.ModTime, "Optional modification unix timestamp override for all files.")
	flag.Var((*appendRegexValue)(&findOpts.Ignore), "ignore", "Regex pattern to ignore")

	flag.Parse()

	if version {
		fmt.Fprintf(os.Stderr, "go-bindata (Go runtime %s).\n", runtime.Version())
		io.WriteString(os.Stderr, "Copyright (c) 2010-2013, Jim Teeuwen.\n")
		io.WriteString(os.Stderr, "Copyright (c) 2017, Tom Thorogood.\n")
		os.Exit(0)
	}

	// Make sure we have input paths.
	if flag.NArg() == 0 {
		io.WriteString(os.Stderr, "Missing <input dir>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if output == "" {
		var err error
		output, err = filepath.Abs("bindata.go")
		must(err)
	}

	genOpts.MemCopy = !noMemCopy
	genOpts.Compress = !noCompress
	genOpts.Metadata = !noMetadata && (genOpts.Mode == 0 || genOpts.ModTime == 0)

	genOpts.Mode = os.FileMode(mode)

	var pkgSet, outputSet bool
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "pkg":
			pkgSet = true
		case "o":
			outputSet = true
		}
	})

	// Change pkg to containing directory of output. If output flag is set and package flag is not.
	if outputSet && !pkgSet {
		pkg := identifier.Identifier(filepath.Base(filepath.Dir(output)))
		if pkg != "" {
			genOpts.Package = pkg
		}
	}

	if !genOpts.MemCopy && genOpts.Compress {
		io.WriteString(os.Stderr, "The use of -nomemcopy without -nocompress is deprecated.\n")
	}

	must(validateOutput(output))
	return
}

func validateOutput(output string) error {
	stat, err := os.Lstat(output)
	if err == nil {
		if stat.IsDir() {
			return errors.New("output path is a directory")
		}

		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	// File does not exist. This is fine, just make
	// sure the directory it is to be in exists.
	if dir, _ := filepath.Split(output); dir != "" {
		return os.MkdirAll(dir, 0744)
	}

	return nil
}

// parseInput determines whether the given path has a recursive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(input string) (path string, recursive bool) {
	return filepath.Clean(strings.TrimSuffix(input, "/...")),
		strings.HasSuffix(input, "/...")
}
