// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/memfs"
)

//
// Watcher is a naive implementation of file event change notification.
//
type Watcher struct {
	// path to file that we want to watch.
	path string

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 millisecond and default is
	// 5 seconds.
	delay time.Duration

	// cb define a function that will be called when file modified or
	// deleted.
	cb WatchCallback

	ticker *time.Ticker
	node   *memfs.Node
}

//
// NewWatcher return a new file watcher that will inspect the file for changes
// with period specified by duration `d` argument.
//
// If duration is less or equal to 100 millisecond, it will be set to default
// duration (5 seconds).
//
func NewWatcher(path string, d time.Duration, cb WatchCallback) (w *Watcher, err error) {
	logp := "NewWatcher"

	if len(path) == 0 {
		return nil, fmt.Errorf("%s: path is empty", logp)
	}
	if cb == nil {
		return nil, fmt.Errorf("%s: callback is not defined", logp)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s: path is directory", logp)
	}

	dummyParent := &memfs.Node{
		SysPath: filepath.Dir(path),
	}
	dummyParent.Path = filepath.Base(dummyParent.SysPath)

	node, err := memfs.NewNode(dummyParent, fi, -1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if d < 100*time.Millisecond {
		d = time.Second * 5
	}

	w = &Watcher{
		path:   path,
		delay:  d,
		cb:     cb,
		ticker: time.NewTicker(d),
		node:   node,
	}

	go w.start()

	return w, nil
}

func (w *Watcher) start() {
	logp := "Watcher"
	if debug.Value >= 2 {
		fmt.Printf("%s: %s: watching for changes\n", logp, w.path)
	}
	for range w.ticker.C {
		ns := &NodeState{
			Node: w.node,
		}

		newInfo, err := os.Stat(w.path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("%s: %s", logp, err.Error())
				continue
			}

			if debug.Value >= 2 {
				fmt.Printf("%s: %s: deleted\n", logp, w.node.SysPath)
			}

			ns.State = FileStateDeleted
			w.cb(ns)
			w.node = nil
			return
		}

		if w.node.Mode() != newInfo.Mode() {
			if debug.Value >= 2 {
				fmt.Printf("%s: %s: mode modified\n", logp, w.node.SysPath)
			}
			ns.State = FileStateUpdateMode
			w.node.SetMode(newInfo.Mode())
			w.cb(ns)
			continue
		}
		if w.node.ModTime().Equal(newInfo.ModTime()) {
			continue
		}
		if debug.Value >= 2 {
			fmt.Printf("%s: %s: content modified\n", logp, w.node.SysPath)
		}

		w.node.SetModTime(newInfo.ModTime())
		w.node.SetSize(newInfo.Size())

		ns.State = FileStateUpdateContent
		w.cb(ns)
	}
}

//
// Stop watching the file.
//
func (w *Watcher) Stop() {
	w.ticker.Stop()
}
