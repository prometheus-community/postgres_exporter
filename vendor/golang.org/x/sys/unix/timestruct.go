// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package unix

// TimespecToNsec converts a Timespec value into a number of
// nanoseconds since the Unix epoch.
func TimespecToNsec(ts Timespec) int64 { return int64(ts.Sec)*1e9 + int64(ts.Nsec) }

// TimevalToNsec converts a Timeval value into a number of nanoseconds
// since the Unix epoch.
func TimevalToNsec(tv Timeval) int64 { return int64(tv.Sec)*1e9 + int64(tv.Usec)*1e3 }
