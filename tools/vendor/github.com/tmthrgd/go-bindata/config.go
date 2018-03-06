// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"errors"
	"hash"
	"os"

	"github.com/tmthrgd/go-bindata/internal/identifier"
)

// HashFormat specifies which format to use when hashing names.
type HashFormat int

const (
	// NameUnchanged leaves the file name unchanged.
	NameUnchanged HashFormat = iota
	// DirHash formats names like path/to/hash/name.ext.
	DirHash
	// NameHashSuffix formats names like path/to/name-hash.ext.
	NameHashSuffix
	// HashWithExt formats names like path/to/hash.ext.
	HashWithExt
)

func (hf HashFormat) String() string {
	switch hf {
	case NameUnchanged:
		return "unchanged"
	case DirHash:
		return "dir"
	case NameHashSuffix:
		return "namesuffix"
	case HashWithExt:
		return "hashext"
	default:
		return "unknown"
	}
}

// HashEncoding specifies which encoding to use when hashing names.
type HashEncoding int

const (
	// HexHash uses hexadecimal encoding.
	HexHash HashEncoding = iota
	// Base32Hash uses unpadded, lowercase standard base32
	// encoding (see RFC 4648).
	Base32Hash
	// Base64Hash uses an unpadded URL-safe base64 encoding
	// defined in RFC 4648.
	Base64Hash
)

func (he HashEncoding) String() string {
	switch he {
	case HexHash:
		return "hex"
	case Base32Hash:
		return "base32"
	case Base64Hash:
		return "base64"
	default:
		return "unknown"
	}
}

// GenerateOptions defines a set of options to use
// when generating the Go code.
type GenerateOptions struct {
	// Name of the package to use.
	Package string

	// Tags specify a set of optional build tags, which should be
	// included in the generated output. The tags are appended to a
	// `// +build` line in the beginning of the output file
	// and must follow the build tags syntax specified by the go tool.
	Tags string

	// MemCopy will alter the way the output file is generated.
	//
	// If false, it will employ a hack that allows us to read the file data directly
	// from the compiled program's `.rodata` section. This ensures that when we call
	// call our generated function, we omit unnecessary mem copies.
	//
	// The downside of this, is that it requires dependencies on the `reflect` and
	// `unsafe` packages. These may be restricted on platforms like AppEngine and
	// thus prevent you from using this mode.
	//
	// Another disadvantage is that the byte slice we create, is strictly read-only.
	// For most use-cases this is not a problem, but if you ever try to alter the
	// returned byte slice, a runtime panic is thrown. Use this mode only on target
	// platforms where memory constraints are an issue.
	//
	// The default behaviour is to use the old code generation method. This
	// prevents the two previously mentioned issues, but will employ at least one
	// extra memcopy and thus increase memory requirements.
	//
	// For instance, consider the following two examples:
	//
	// This would be the default mode, using an extra memcopy but gives a safe
	// implementation without dependencies on `reflect` and `unsafe`:
	//
	// 	func myfile() []byte {
	// 		return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a}
	// 	}
	//
	// Here is the same functionality, but uses the `.rodata` hack.
	// The byte slice returned from this example can not be written to without
	// generating a runtime error.
	//
	// 	var _myfile = "\x89\x50\x4e\x47\x0d\x0a\x1a"
	//
	// 	func myfile() []byte {
	// 		var empty [0]byte
	// 		sx := (*reflect.StringHeader)(unsafe.Pointer(&_myfile))
	// 		b := empty[:]
	// 		bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	// 		bx.Data = sx.Data
	// 		bx.Len = len(_myfile)
	// 		bx.Cap = bx.Len
	// 		return b
	// 	}
	MemCopy bool

	// Compress means the assets are GZIP compressed before being turned into
	// Go code. The generated function will automatically unzip the file data
	// when called. Defaults to true.
	Compress bool

	// Perform a debug build. This generates an asset file, which
	// loads the asset contents directly from disk at their original
	// location, instead of embedding the contents in the code.
	//
	// This is mostly useful if you anticipate that the assets are
	// going to change during your development cycle. You will always
	// want your code to access the latest version of the asset.
	// Only in release mode, will the assets actually be embedded
	// in the code. The default behaviour is Release mode.
	Debug bool

	// Perform a dev build, which is nearly identical to the debug option. The
	// only difference is that instead of absolute file paths in generated code,
	// it expects a variable, `rootDir`, to be set in the generated code's
	// package (the author needs to do this manually), which it then prepends to
	// an asset's name to construct the file path on disk.
	//
	// This is mainly so you can push the generated code file to a shared
	// repository.
	Dev bool

	// When true, the AssetDir API will be provided.
	AssetDir bool

	// When true, only gzip decompress the data on first use.
	DecompressOnce bool

	// [Deprecated]: use github.com/tmthrgd/go-bindata/restore.
	Restore bool

	// When false, size, mode and modtime are not preserved from files
	Metadata bool
	// When nonzero, use this as mode for all files.
	Mode os.FileMode
	// When nonzero, use this as unix timestamp for all files.
	ModTime int64

	// Hash is used to produce a hash of the file.
	Hash hash.Hash
	// Which of the given name hashing formats to use.
	HashFormat HashFormat
	// The length of the hash to use, defaults to 16 characters.
	HashLength uint
	// The encoding to use to encode the name hash.
	HashEncoding HashEncoding
}

// validate ensures the config has sane values.
// Part of which means checking if certain file/directory paths exist.
func (opts *GenerateOptions) validate() error {
	if len(opts.Package) == 0 {
		return errors.New("go-bindata: missing package name")
	}

	if identifier.Identifier(opts.Package) != opts.Package {
		return errors.New("go-bindata: package name is not valid identifier")
	}

	if opts.Metadata && (opts.Mode != 0 && opts.ModTime != 0) {
		return errors.New("go-bindata: if Metadata is true, one of Mode or ModTime must be zero")
	}

	if opts.Mode&^os.ModePerm != 0 {
		return errors.New("go-bindata: invalid mode specified")
	}

	if opts.Hash != nil && (opts.Debug || opts.Dev) {
		return errors.New("go-bindata: Hash is not compatible with Debug and Dev")
	}

	if opts.Restore && !opts.AssetDir {
		return errors.New("go-bindata: Restore cannot be used without AssetDir")
	}

	return nil
}
