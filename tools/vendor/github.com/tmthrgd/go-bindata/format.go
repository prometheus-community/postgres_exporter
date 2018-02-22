// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
)

var printerConfig = printer.Config{
	Mode:     printer.UseSpaces | printer.TabIndent,
	Tabwidth: 8,
}

func formatTemplate(name string, data interface{}) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	buf.WriteString("package main;")

	if err := baseTemplate.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", buf, parser.ParseComments)
	if err != nil {
		return "", err
	}

	buf.Reset()

	if err = printerConfig.Fprint(buf, fset, f); err != nil {
		return "", err
	}

	out := string(bytes.TrimSpace(buf.Bytes()[len("package main\n"):]))
	return out, nil
}
