// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"os"
	"testing"
)

// 999f056 With bytes.Buffer in functions
// BenchmarkParse-2  300   17007586 ns/op  6361586 B/op  78712 allocs/op/
//
// 22dcd07 Move buffer to reader
// BenchmarkParse-2  500   19534400 ns/op  4656335 B/op  81163 allocs/op
//
// 488e6c3 Refactor parser using bytes.Reader
// BenchmarkParse-2  20000    72338 ns/op    35400 B/op    550 allocs/op
//
// Replace field type in Section and Variable from []byte to string
// BenchmarkParse-2  20000    61150 ns/op    25176 B/op    482 allocs/op
func BenchmarkParse(b *testing.B) {
	reader := newReader()
	src, err := os.ReadFile(testdataInputIni)
	if err != nil {
		b.Fatal(err)
	}

	for x := 0; x < b.N; x++ {
		_, _ = reader.Parse(src)
	}
}
