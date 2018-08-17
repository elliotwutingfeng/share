// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
)

var rrPool = sync.Pool{
	New: func() interface{} {
		rr := &ResourceRecord{
			Text: &RDataText{},
		}
		return rr
	},
}