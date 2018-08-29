// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package main

import (
	"bytes"
	"regexp"
)

type appendRegexValue []*regexp.Regexp

func (ar *appendRegexValue) String() string {
	if ar == nil {
		return ""
	}

	var buf bytes.Buffer

	for i, r := range *ar {
		if i != 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(r.String())
	}

	return buf.String()
}

func (ar *appendRegexValue) Set(value string) error {
	r, err := regexp.Compile(value)
	if err != nil {
		return err
	}

	if *ar == nil {
		*ar = make([]*regexp.Regexp, 0, 1)
	}

	*ar = append(*ar, r)
	return nil
}
