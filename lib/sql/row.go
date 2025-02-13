// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"fmt"
	"sort"
)

// Row represent a column-name and value in a tuple.
// The map's key is the column name in database and the map's value is
// the column's value.
// This type can be used to create dynamic insert-update fields.
type Row map[string]interface{}

// ExtractSQLFields extract the column's name, column place holder, and column
// values as slices.
//
// The driverName define the returned place holders.
// If the driverName is "postgres" then the list of holders will be returned
// as counter, for example "$1", "$2" and so on.
// If the driverName is "mysql" or empty or unknown the the list of holders
// will be returned as list of "?".
//
// The returned names will be sorted in ascending order.
func (row Row) ExtractSQLFields(driverName string) (names, holders []string, values []interface{}) {
	if len(row) == 0 {
		return nil, nil, nil
	}

	names = make([]string, 0, len(row))
	holders = make([]string, 0, len(row))
	values = make([]interface{}, 0, len(row))

	for k := range row {
		names = append(names, k)
	}
	sort.Strings(names)

	for x, k := range names {
		if driverName == DriverNamePostgres {
			holders = append(holders, fmt.Sprintf("$%d", x+1))
		} else {
			holders = append(holders, DefaultPlaceHolder)
		}
		values = append(values, row[k])
	}

	return names, holders, values
}
