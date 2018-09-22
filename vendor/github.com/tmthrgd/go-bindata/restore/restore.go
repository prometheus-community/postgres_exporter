// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

// Package restore provides the restore API that was
// previously embedded into the generated output.
package restore

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// AssetAndInfo represents the generated AssetAndInfo method.
type AssetAndInfo func(name string) (data []byte, info os.FileInfo, err error)

// AssetDir represents the generated AssetDir method.
type AssetDir func(name string) (children []string, err error)

// Asset restores an asset under the given directory
func Asset(dir, name string, assetAndInfo AssetAndInfo) error {
	path := filepath.Join(dir, filepath.FromSlash(name))

	data, info, err := assetAndInfo(name)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err = ioutil.WriteFile(path, data, info.Mode()); err != nil {
		return err
	}

	return os.Chtimes(path, info.ModTime(), info.ModTime())
}

// Assets restores an asset under the given directory recursively
func Assets(dir, name string, assetDir AssetDir, assetAndInfo AssetAndInfo) error {
	children, err := assetDir(name)
	// File
	if err != nil {
		return Asset(dir, name, assetAndInfo)
	}

	// Dir
	for _, child := range children {
		if err = Assets(dir, filepath.Join(name, child), assetDir, assetAndInfo); err != nil {
			return err
		}
	}

	return nil
}
