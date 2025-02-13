// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"strings"
	"time"
)

const (
	defaultRName      string = "root"
	defaultRefresh    int32  = 1 * 24 * 60 * 60 // 1 Day.
	defaultRetry      int32  = 1 * 60 * 60      // 1 Hour.
	defaultMinimumTTL uint32 = 1 * 60 * 60      // 1 Hour.
)

// RDataSOA contains the authority of zone.
//
// All times are in units of seconds.
type RDataSOA struct {
	// The primary name server for the zone.
	MName string

	// The mailbox of the person responsible for the name server.
	// For example, "root@localhost", but with '@' is replaced with dot
	// '.', its become "root.localhost".
	// If its empty, default to "root".
	RName string

	// The version number of the original copy of the zone.
	// If its empty, default to current epoch.
	Serial uint32

	// A time interval before the zone should be refreshed.
	// If its empty, default to 1 days.
	Refresh int32

	// A time interval that should elapse before a failed refresh should
	// be retried.
	// If its empty, default to 1 hour.
	Retry int32

	// A time value that specifies the upper limit on the time interval
	// that can elapse before the zone is no longer authoritative.
	Expire int32

	// The minimum TTL field that should be exported with any RR from this
	// zone.
	// If its empty, default to 1 hour.
	Minimum uint32
}

// init initialize the SOA record by setting fields to its default value if
// its empty.
func (soa *RDataSOA) init() {
	if len(soa.MName) > 0 {
		soa.MName = strings.ToLower(soa.MName)
	}
	if len(soa.RName) > 0 {
		soa.RName = strings.ToLower(soa.RName)
	} else {
		soa.RName = defaultRName
	}
	if soa.Serial == 0 {
		soa.Serial = uint32(time.Now().Unix())
	}
	if soa.Refresh == 0 {
		soa.Refresh = defaultRefresh
	}
	if soa.Retry == 0 {
		soa.Retry = defaultRetry
	}
	if soa.Minimum == 0 {
		soa.Minimum = defaultMinimumTTL
	}
}
