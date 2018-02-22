// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

// +build !go1.9

package bindata

import (
	"encoding/base32"
	"strings"
)

var base32Enc = base32EncodingCompat{
	base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567"),
}

type base32EncodingCompat struct{ *base32.Encoding }

func (enc base32EncodingCompat) EncodeToString(src []byte) string {
	return strings.TrimSuffix(enc.Encoding.EncodeToString(src), "=")
}
