// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package identifier

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Identifier removes all characters from a string that are not valid in
// an identifier according to the Go Programming Language Specification.
//
// The logic in the switch statement was taken from go/source package:
// https://github.com/golang/go/blob/a1a688fa0012f7ce3a37e9ac0070461fe8e3f28e/src/go/scanner/scanner.go#L257-#L271
func Identifier(val string) string {
	return strings.TrimLeftFunc(strings.Map(func(ch rune) rune {
		switch {
		case 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
			ch >= utf8.RuneSelf && unicode.IsLetter(ch):
			return ch
		case '0' <= ch && ch <= '9' ||
			ch >= utf8.RuneSelf && unicode.IsDigit(ch):
			return ch
		default:
			return -1
		}
	}, val), unicode.IsDigit)
}
