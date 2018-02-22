// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"encoding/base64"
	"encoding/hex"
	"path"
	"strings"
)

// Name applies name hashing if required. It returns the original
// name for NoHash and NameUnchanged and returns the mangledName
// otherwise.
func (asset *binAsset) Name() string {
	if asset.Hash == nil || asset.opts.HashFormat == NameUnchanged {
		return asset.File.Name()
	} else if asset.mangledName != "" {
		return asset.mangledName
	}

	var enc string
	switch asset.opts.HashEncoding {
	case HexHash:
		enc = hex.EncodeToString(asset.Hash)
	case Base32Hash:
		enc = base32Enc.EncodeToString(asset.Hash)
	case Base64Hash:
		enc = base64.RawURLEncoding.EncodeToString(asset.Hash)
	default:
		panic("unreachable")
	}

	l := asset.opts.HashLength
	if l == 0 {
		l = 16
	}

	if l < uint(len(enc)) {
		enc = enc[:l]
	}

	dir, file := path.Split(asset.File.Name())
	ext := path.Ext(file)

	switch asset.opts.HashFormat {
	case DirHash:
		asset.mangledName = path.Join(dir, enc, file)
	case NameHashSuffix:
		file = strings.TrimSuffix(file, ext)
		asset.mangledName = path.Join(dir, file+"-"+enc+ext)
	case HashWithExt:
		asset.mangledName = path.Join(dir, enc+ext)
	default:
		panic("unreachable")
	}

	return asset.mangledName
}
