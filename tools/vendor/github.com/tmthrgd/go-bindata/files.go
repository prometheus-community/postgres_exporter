// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// File represents a single asset file.
type File interface {
	// Name returns the name by which asset is referenced.
	Name() string
	// Path returns the relative path to the file.
	Path() string
	// AbsolutePath returns the absolute path to the file.
	AbsolutePath() string

	// Open returns an io.ReadCloser for reading the file.
	Open() (io.ReadCloser, error)
	// Stat returns an os.FileInfo interface representing the file.
	Stat() (os.FileInfo, error)
}

// Files represents a collection of asset files.
type Files []File

type osFile struct {
	name string
	path string
}

func (f *osFile) Name() string {
	return f.name
}

func (f *osFile) Path() string {
	return f.path
}

func (f *osFile) AbsolutePath() string {
	path, err := filepath.Abs(f.path)
	if err != nil {
		return f.path
	}

	return path
}

func (f *osFile) Open() (io.ReadCloser, error) {
	return os.Open(f.path)
}

func (f *osFile) Stat() (os.FileInfo, error) {
	return os.Stat(f.path)
}

// FindFilesOptions defines a set of options to use
// when searching for files.
type FindFilesOptions struct {
	// Prefix defines a path prefix which should be stripped from all
	// file names when generating the keys in the table of contents.
	// For example, running without the `-prefix` flag, we get:
	//
	// 	$ go-bindata /path/to/templates
	// 	go_bindata["/path/to/templates/foo.html"] = _path_to_templates_foo_html
	//
	// Running with the `-prefix` flag, we get:
	//
	// 	$ go-bindata -prefix "/path/to/" /path/to/templates/foo.html
	// 	go_bindata["templates/foo.html"] = templates_foo_html
	Prefix string

	// Recursive defines whether subdirectories of Path
	// should be recursively included in the conversion.
	Recursive bool

	// Ignores any filenames matching the regex pattern specified, e.g.
	// path/to/file.ext will ignore only that file, or \\.gitignore
	// will match any .gitignore file.
	//
	// This parameter can be provided multiple times.
	Ignore []*regexp.Regexp
}

// FindFiles adds all files inside a directory to the
// generated output.
func FindFiles(path string, opts *FindFilesOptions) (files Files, err error) {
	if opts == nil {
		opts = new(FindFilesOptions)
	}

	if err = filepath.Walk(path, func(assetPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !opts.Recursive && assetPath != path {
				return filepath.SkipDir
			}

			return nil
		}

		for _, re := range opts.Ignore {
			if re.MatchString(assetPath) {
				return nil
			}
		}

		name := strings.TrimPrefix(filepath.ToSlash(
			strings.TrimPrefix(assetPath, opts.Prefix)), "/")
		if name == "" {
			panic("should be impossible")
		}

		files = append(files, &osFile{name, assetPath})
		return nil
	}); err != nil {
		return nil, err
	}

	return
}
