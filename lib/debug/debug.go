// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package debug provide global debug variable, initialized through
// environment variable "DEBUG" or directly.
package debug

import (
	"os"
	"strconv"
)

var (
	// Value contains DEBUG value from environment.
	Value = 0
)

//
// init initialize debug from system environment.
//
func init() {
	v := os.Getenv("DEBUG")
	if len(v) > 0 {
		Value, _ = strconv.Atoi(v)
	}
}
