// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"bytes"
	"compress/flate"
	"io"
	"path"
	"strings"
	"sync"
	"text/template"
)

var flatePool sync.Pool

func writeWrappedString(write func(io.Writer) error, indent string, wrapAt int) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	buf.WriteString("(\"\" +\n")
	buf.WriteString(indent)
	buf.WriteByte('"')

	if err := write(&stringWriter{
		Writer: buf,
		Indent: indent,
		WrapAt: wrapAt,
	}); err != nil {
		return "", err
	}

	buf.WriteString("\")")

	s := buf.String()

	if strings.IndexByte(s[1:], '(') == -1 {
		s = s[1 : len(s)-1]
	}

	return s, nil
}

func init() {
	template.Must(template.Must(baseTemplate.New("release").Funcs(template.FuncMap{
		"base": path.Base,
		"wrap": func(data []byte, indent string, wrapAt int) (string, error) {
			return writeWrappedString(func(w io.Writer) error {
				_, err := w.Write(data)
				return err
			}, indent, wrapAt)
		},
		"read": func(asset binAsset, indent string, wrapAt int) (string, error) {
			return writeWrappedString(asset.copy, indent, wrapAt)
		},
		"flate": func(asset binAsset, indent string, wrapAt int) (out string, err error) {
			return writeWrappedString(func(w io.Writer) error {
				fw, _ := flatePool.Get().(*flate.Writer)
				if fw != nil {
					fw.Reset(w)
				} else if fw, err = flate.NewWriter(w, flate.BestCompression); err != nil {
					return err
				}
				defer flatePool.Put(fw)

				if err := asset.copy(fw); err != nil {
					return err
				}

				return fw.Close()
			}, indent, wrapAt)
		},
		"format": formatTemplate,
	}).Parse(`
{{- $unsafeRead := and (not $.Compress) (not $.MemCopy) -}}
import (
{{- if $.Compress}}
	"bytes"
	"compress/flate"
	"io"
{{- end}}
	"os"
	"path/filepath"
{{- if $unsafeRead}}
	"reflect"
{{- end}}
{{- if or $.Compress $.AssetDir}}
	"strings"
{{- end}}
{{- if and $.Compress $.DecompressOnce}}
	"sync"
{{- end}}
	"time"
{{- if $unsafeRead}}
	"unsafe"
{{- end}}
{{- if $.Restore}}

	"github.com/tmthrgd/go-bindata/restore"
{{- end}}
)

{{if $unsafeRead -}}
func bindataRead(data string) []byte {
	var empty [0]byte
	sx := (*reflect.StringHeader)(unsafe.Pointer(&data))
	b := empty[:]
	bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bx.Data = sx.Data
	bx.Len = len(data)
	bx.Cap = bx.Len
	return b
}

{{end -}}

type asset struct {
	name string
{{- if and $.Hash $.HashFormat}}
	orig string
{{- end}}
	data string
{{- if $.Compress}}
	size int64
{{- end -}}
{{- if and $.Metadata (not $.Mode)}}
	mode os.FileMode
{{- end -}}
{{- if and $.Metadata (not $.ModTime)}}
	time time.Time
{{- end -}}
{{- if $.Hash}}
	hash string
{{- end}}
{{- if and $.Compress $.DecompressOnce}}

	once  sync.Once
	bytes []byte
	err   error
{{- end}}
}

func (a *asset) Name() string {
	return a.name
}

func (a *asset) Size() int64 {
{{- if $.Compress}}
	return a.size
{{- else}}
	return int64(len(a.data))
{{- end}}
}

func (a *asset) Mode() os.FileMode {
{{- if $.Mode}}
	return {{printf "%04o" $.Mode}}
{{- else if $.Metadata}}
	return a.mode
{{- else}}
	return 0
{{- end}}
}

func (a *asset) ModTime() time.Time {
{{- if $.ModTime}}
	return time.Unix({{$.ModTime}}, 0)
{{- else if $.Metadata}}
	return a.time
{{- else}}
	return time.Time{}
{{- end}}
}

func (*asset) IsDir() bool {
	return false
}

func (*asset) Sys() interface{} {
	return nil
}

{{- if $.Hash}}

func (a *asset) OriginalName() string {
{{- if $.HashFormat}}
	return a.orig
{{- else}}
	return a.name
{{- end}}
}

func (a *asset) FileHash() []byte {
{{- if $unsafeRead}}
	return bindataRead(a.hash)
{{- else}}
	return []byte(a.hash)
{{- end}}
}

type FileInfo interface {
	os.FileInfo

	OriginalName() string
	FileHash() []byte
}
{{- end}}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]*asset{
{{range $.Assets}}	{{printf "%q" .Name}}: &asset{
		name: {{printf "%q" (base .Name)}},
	{{- if and $.Hash $.HashFormat}}
		orig: {{printf "%q" .File.Name}},
	{{- end}}
		data: {{if $.Compress -}}
			{{flate . "\t\t\t" 24}}
		{{- else -}}
			{{read . "\t\t\t" 24}}
		{{- end}},

	{{- if or $.Metadata $.Compress -}}
		{{- $info := .Stat -}}

		{{- if $.Compress}}
		size: {{$info.Size}},
		{{- end -}}

		{{- if and $.Metadata (not $.Mode)}}
		mode: {{printf "%04o" $info.Mode}},
		{{- end -}}

		{{- if and $.Metadata (not $.ModTime)}}
		{{$mod := $info.ModTime -}}
		time: time.Unix({{$mod.Unix}}, {{$mod.Nanosecond}}),
		{{- end -}}
	{{- end -}}

	{{- if $.Hash}}
		hash: {{wrap .Hash "\t\t\t" 24}},
	{{- end}}
	},
{{end -}}
}

// AssetAndInfo loads and returns the asset and asset info for the
// given name. It returns an error if the asset could not be found
// or could not be loaded.
func AssetAndInfo(name string) ([]byte, os.FileInfo, error) {
	a, ok := _bindata[filepath.ToSlash(name)]
	if !ok {
		return nil, nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
{{if and $.Compress $.DecompressOnce}}
	a.once.Do(func() {
		fr := flate.NewReader(strings.NewReader(a.data))

		var buf bytes.Buffer
		if _, a.err = io.Copy(&buf, fr); a.err != nil {
			return
		}

		if a.err = fr.Close(); a.err == nil {
			a.bytes = buf.Bytes()
		}
	})
	if a.err != nil {
		return nil, nil, &os.PathError{Op: "read", Path: name, Err: a.err}
	}

	return a.bytes, a, nil
{{- else if $.Compress}}
	fr := flate.NewReader(strings.NewReader(a.data))

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, fr); err != nil {
		return nil, nil, &os.PathError{Op: "read", Path: name, Err: err}
	}

	if err := fr.Close(); err != nil {
		return nil, nil, &os.PathError{Op: "read", Path: name, Err: err}
	}

	return buf.Bytes(), a, nil
{{- else if $unsafeRead}}
	return bindataRead(a.data), a, nil
{{- else}}
	return []byte(a.data), a, nil
{{- end}}
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	a, ok := _bindata[filepath.ToSlash(name)]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	return a, nil
}

{{- if and $.Hash $.HashFormat}}

{{format "hashnames" $}}

// AssetName returns the hashed name associated with an asset of a
// given name.
func AssetName(name string) (string, error) {
	if name, ok := _hashNames[filepath.ToSlash(name)]; ok {
		return name, nil
	}

	return "", &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}
{{- end}}`)).New("hashnames").Parse(`
var _hashNames = map[string]string{
{{range .Assets -}}
	{{printf "%q" .File.Name}}: {{printf "%q" .Name}},
{{end -}}
}`))
}
