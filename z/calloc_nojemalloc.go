// Copyright 2020 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

// +build !jemalloc

package z

import (
	"fmt"
	"sync/atomic"
)

// Provides versions of New and Free when cgo is not available (e.g. cross
// compilation).

// Calloc allocates a slice of size n.
func Calloc(n int) []byte {
	atomic.AddInt64(&numBytes, int64(n))
	return make([]byte, n)
}

// CallocNoRef will not give you memory back without jemalloc.
func CallocNoRef(n int) []byte {
	return nil
}

// Free does not do anything in this mode.
func Free(b []byte) {
	atomic.AddInt64(&numBytes, -int64(cap(b)))
}

func StatsPrint() {
	fmt.Printf("Using Go memory")
}
