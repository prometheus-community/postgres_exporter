// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import "text/template"

func init() {
	template.Must(baseTemplate.New("common").Parse(`// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	data, _, err := AssetAndInfo(name)
	return data, err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}

	return names
}

{{- if $.Restore}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	return restore.Asset(dir, name, AssetAndInfo)
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	return restore.Assets(dir, name, AssetDir, AssetAndInfo)
}
{{- end}}`))
}
