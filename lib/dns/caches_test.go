// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestCachesGet(t *testing.T) {
	type testCase struct {
		exp     *Answer
		desc    string
		QName   string
		expList []*Answer
		RType   RecordType
		RClass  RecordClass
	}

	var (
		an1 = &Answer{
			ReceivedAt: 1,
			QName:      "test",
			RType:      1,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 1,
				},
			},
		}
		an2 = &Answer{
			ReceivedAt: 2,
			QName:      "test",
			RType:      2,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 2,
				},
			},
		}
		an3 = &Answer{
			ReceivedAt: 3,
			QName:      "test",
			RType:      3,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 3,
				},
			},
		}

		ca      Caches
		cases   []testCase
		c       testCase
		got     *Answer
		gotList []*Answer
	)

	ca.init(0, 0, 0)
	ca.upsert(an1)
	ca.upsert(an2)
	ca.upsert(an3)

	cases = []testCase{{
		desc: "With query not found",
		expList: []*Answer{
			an1, an2, an3,
		},
	}, {
		desc:   "With query found",
		QName:  "test",
		RType:  1,
		RClass: 1,
		exp:    an1,
		expList: []*Answer{
			an2, an3, an1,
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		_, got = ca.get(c.QName, c.RType, c.RClass)
		gotList = ca.ExternalLRU()

		test.Assert(t, "caches.get", c.exp, got)
		test.Assert(t, "caches.list", c.expList, gotList)
	}
}

func TestCachesPrune(t *testing.T) {
	type testCase struct {
		desc    string
		expList []*Answer
	}

	var (
		at = time.Now().Unix()

		an1 = &Answer{
			ReceivedAt: 1,
			AccessedAt: 1,
			QName:      "test",
			RType:      1,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 1,
				},
			},
		}
		an2 = &Answer{
			ReceivedAt: 2,
			AccessedAt: 2,
			QName:      "test",
			RType:      2,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 2,
				},
			},
		}
		an3 = &Answer{
			ReceivedAt: at,
			AccessedAt: at,
			QName:      "test",
			RType:      3,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 3,
				},
			},
		}

		ca      Caches
		cases   []testCase
		c       testCase
		gotList []*Answer
	)

	ca.init(0, 0, 0)
	ca.upsert(an1)
	ca.upsert(an2)
	ca.upsert(an3)

	t.Logf("%+v\n", ca.ExternalLRU())

	cases = []testCase{{
		desc: "With several caches got pruned",
		expList: []*Answer{
			an3,
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		_ = ca.prune(3)

		gotList = ca.ExternalLRU()

		test.Assert(t, "caches.list", c.expList, gotList)
	}
}

func TestCaches_ExternalSave(t *testing.T) {
	var (
		srv = &Server{}

		hname   = []byte("caches.save.local")
		address = []byte("127.0.0.1")
		msg     = NewMessageAddress(hname, [][]byte{address})
		answer  = newAnswer(msg, false)

		w          bytes.Buffer
		expAnswers []*Answer
		gotAnswers []*Answer
		err        error
		n          int
	)

	srv.Caches.init(0, 0, 0)

	_ = srv.Caches.upsert(answer)

	n, err = srv.Caches.ExternalSave(&w)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Caches.ExternalSave", 1, n)

	msg = NewMessage()
	msg.packet = answer.msg.packet
	err = msg.Unpack()
	if err != nil {
		t.Fatal(err)
	}
	expAnswers = append(expAnswers, newAnswer(msg, false))

	srv.Caches.init(0, 0, 0)

	gotAnswers, err = srv.Caches.ExternalLoad(&w)
	if err != nil {
		t.Fatal(err)
	}

	for _, answer = range gotAnswers {
		answer.el = nil
	}

	test.Assert(t, "Caches.Write", expAnswers, gotAnswers)
}

func TestCachesUpsert(t *testing.T) {
	type testCase struct {
		nu      *Answer
		desc    string
		expList []*Answer
		expLen  int
	}

	var (
		an1 = &Answer{
			ReceivedAt: 1,
			AccessedAt: 1,
			QName:      "test",
			RType:      1,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 1,
				},
			},
		}
		an1Update = &Answer{
			ReceivedAt: 3,
			AccessedAt: 3,
			QName:      "test",
			RType:      1,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 3,
				},
			},
		}
		an2 = &Answer{
			ReceivedAt: 2,
			AccessedAt: 2,
			QName:      "test",
			RType:      2,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 2,
				},
			},
		}
		an2Update = &Answer{
			ReceivedAt: 4,
			AccessedAt: 4,
			QName:      "test",
			RType:      2,
			RClass:     1,
			msg: &Message{
				Header: MessageHeader{
					ID: 4,
				},
			},
		}

		ca      Caches
		cases   []testCase
		c       testCase
		gotList []*Answer
		x       int
	)

	ca.init(0, 0, 0)

	cases = []testCase{{
		desc: "With empty answer",
	}, {
		desc:    "With new answer",
		nu:      an1,
		expLen:  1,
		expList: []*Answer{an1},
	}, {
		desc:    "With new answer, different type",
		nu:      an2,
		expLen:  2,
		expList: []*Answer{an1, an2},
	}, {
		desc:    "With update on answer",
		nu:      an1Update,
		expLen:  2,
		expList: []*Answer{an1, an2},
	}, {
		desc:    "With update on answer (2)",
		nu:      an2Update,
		expLen:  2,
		expList: []*Answer{an1, an2},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		ca.upsert(c.nu)

		gotList = ca.ExternalLRU()

		test.Assert(t, "len(caches.list)", c.expLen, len(gotList))

		for x = 0; x < len(gotList); x++ {
			test.Assert(t, "caches.list", c.expList[x], gotList[x])
		}
	}
}
