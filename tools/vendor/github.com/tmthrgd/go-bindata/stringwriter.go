// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import "io"

var (
	stringWriterLinePrefix       = []byte(`"`)
	stringWriterLineSuffix       = []byte("\" +\n")
	stringWriterParensLineSuffix = []byte("\") + (\"\" +\n")
)

type stringWriter struct {
	io.Writer
	Indent string
	WrapAt int
	c, l   int
}

func (w *stringWriter) Write(p []byte) (n int, err error) {
	buf := [4]byte{'\\', 'x', 0, 0}

	for _, b := range p {
		const lowerHex = "0123456789abcdef"
		buf[2] = lowerHex[b/16]
		buf[3] = lowerHex[b%16]

		if _, err = w.Writer.Write(buf[:]); err != nil {
			return
		}

		n++
		w.c++

		if w.WrapAt == 0 || w.c%w.WrapAt != 0 {
			continue
		}

		w.l++

		suffix := stringWriterLineSuffix
		if w.l%500 == 0 {
			// As per https://golang.org/issue/18078, the compiler has trouble
			// compiling the concatenation of many strings, s0 + s1 + s2 + ... + sN,
			// for large N. We insert redundant, explicit parentheses to work around
			// that, lowering the N at any given step: (s0 + s1 + ... + s499) + (s500 +
			// ... + s1999) + etc + (etc + ... + sN).
			//
			// This fix was taken from the fix applied to x/text in
			// https://github.com/golang/text/commit/5c6cf4f9a2.

			suffix = stringWriterParensLineSuffix
		}

		if _, err = w.Writer.Write(suffix); err != nil {
			return
		}

		if _, err = io.WriteString(w.Writer, w.Indent); err != nil {
			return
		}

		if _, err = w.Writer.Write(stringWriterLinePrefix); err != nil {
			return
		}
	}

	return
}
