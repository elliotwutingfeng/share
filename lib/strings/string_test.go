// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestCleanURI(t *testing.T) {
	cases := []struct {
		text string
		exp  string
	}{{
		// Empty
	}, {
		text: `ftp://test.com/123 The [[United States]] has regularly voted alone and against international consensus, using its [[United Nations Security Council veto power|veto power]] to block the adoption of proposed UN Security Council resolutions supporting the [[PLO]] and calling for a two-state solution to the [[Israeli-Palestinian conflict]].<ref>[http://books.google.ca/books?id=CHL5SwGvobQC&pg=PA168&dq=US+veto+Israel+regularly#v=onepage&q=US%20veto%20Israel%20regularly&f=false Pirates and emperors, old and new: international terrorism in the real world], [[Noam Chomsky]], p. 168.</ref><ref>The US has also used its veto to block resolutions that are critical of Israel.[https://books.google.ca/books?id=yzmpDAz7ZAwC&pg=PT251&dq=US+veto+Israel+regularly&lr=#v=onepage&q=US%20veto%20Israel%20regularly&f=false Uneasy neighbors], David T. Jones and David Kilgour, p. 235.</ref> The United States responded to the frequent criticism from UN organs by adopting the [[Negroponte doctrine]].`,
		exp:  ` The [[United States]] has regularly voted alone and against international consensus, using its [[United Nations Security Council veto power|veto power]] to block the adoption of proposed UN Security Council resolutions supporting the [[PLO]] and calling for a two-state solution to the [[Israeli-Palestinian conflict]].<ref>[ Pirates and emperors, old and new: international terrorism in the real world], [[Noam Chomsky]], p. 168.</ref><ref>The US has also used its veto to block resolutions that are critical of Israel.[ Uneasy neighbors], David T. Jones and David Kilgour, p. 235.</ref> The United States responded to the frequent criticism from UN organs by adopting the [[Negroponte doctrine]].`,
	}}

	for _, c := range cases {
		got := CleanURI(c.text)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCleanWikiMarkup(t *testing.T) {
	cases := []struct {
		text string
		exp  string
	}{{
		text: `==External links==
*[http://www.bigfinish.com/24-Doctor-Who-The-Eye-of-the-Scorpion Big Finish Productions - ''The Eye of the Scorpion'']
*{{Doctor Who RG | id=who_bf24 | title=The Eye of the Scorpion}}
===Reviews===
* Test image [[Image:fileto.png]].
* Test file [[File:fileto.png]].
*{{OG review | id=bf-24 | title=The Eye of the Scorpion}}
*{{DWRG | id=eyes | title=The Eye of the Scorpion}}

<br clear="all">
{{Fifthdoctoraudios}}

{{DEFAULTSORT:Eye of the Scorpion, The}}
[[Category:Fifth Doctor audio plays]]
[[Category:Fifth Doctor audio plays]]
[[:Category:2001 audio plays]]
{{DoctorWho-stub}}`,
		exp: `==External links==
*[http://www.bigfinish.com/24-Doctor-Who-The-Eye-of-the-Scorpion Big Finish Productions - ''The Eye of the Scorpion'']
*{{Doctor Who RG | id=who_bf24 | title=The Eye of the Scorpion}}
===Reviews===
* Test image .
* Test file .
*{{OG review | id=bf-24 | title=The Eye of the Scorpion}}
*{{DWRG | id=eyes | title=The Eye of the Scorpion}}

<br clear="all">
{{Fifthdoctoraudios}}





{{DoctorWho-stub}}`,
	}}

	for _, c := range cases {
		got := CleanWikiMarkup(c.text)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestMergeSpaces(t *testing.T) {
	cases := []struct {
		text     string
		withline bool
		exp      string
	}{{
		text: "   a\n\nb c   d\n\n",
		exp:  " a\n\nb c d\n\n",
	}, {
		text: " \t a \t ",
		exp:  " a ",
	}, {
		text:     "   a\n\nb c   d\n\n",
		withline: true,
		exp:      " a\nb c d\n",
	}}

	for _, c := range cases {
		got := MergeSpaces(c.text, c.withline)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestReverse(t *testing.T) {
	cases := []struct {
		input string
		exp   string
	}{{
		input: "The quick bròwn 狐 jumped over the lazy 犬",
		exp:   "犬 yzal eht revo depmuj 狐 nwòrb kciuq ehT",
	}}

	for _, c := range cases {
		got := Reverse(c.input)

		test.Assert(t, "Reverse", c.exp, got, true)
	}
}

func TestSingleSpace(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in: "",
	}, {
		in:  " \t\v\r\n\r\n\fa \t\v\r\n\r\n\f",
		exp: " a ",
	}}
	for _, c := range cases {
		got := SingleSpace(c.in)
		test.Assert(t, c.in, c.exp, got, true)
	}
}

func TestSplit(t *testing.T) {
	cases := []struct {
		text string
		exp  []string
	}{{
		text: `// Copyright 2016-2018 Shulhan <ms@kilabit.info>. All rights reserved.`,
		exp: []string{"Copyright", "2016-2018", "Shulhan",
			"ms@kilabit.info", "All", "rights", "reserved"},
	}, {
		text: `The [[United States]] has regularly voted alone and
		against international consensus, using its [[United Nations
		Security Council veto power|veto power]] to block the adoption
		of proposed UN Security Council resolutions supporting the
		[[PLO]] and calling for a two-state solution to the
		[[Israeli-Palestinian conflict]].`,
		exp: []string{"The", "United", "States", "has", "regularly",
			"voted", "alone", "and", "against", "international",
			"consensus", "using", "its", "Nations", "Security",
			"Council", "veto", "power|veto", "power", "to",
			"block", "adoption", "of", "proposed", "UN",
			"resolutions", "supporting", "PLO", "calling",
			"for", "a", "two-state", "solution",
			"Israeli-Palestinian", "conflict",
		},
	}}

	for _, c := range cases {
		got := Split(c.text, true, true)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestTrimNonAlnum(t *testing.T) {
	cases := []struct {
		text string
		exp  string
	}{
		{"[[alpha]]", "alpha"},
		{"[[alpha", "alpha"},
		{"alpha]]", "alpha"},
		{"alpha", "alpha"},
		{"alpha0", "alpha0"},
		{"1alpha", "1alpha"},
		{"1alpha0", "1alpha0"},
		{"[][][]", ""},
	}

	for _, c := range cases {
		got := TrimNonAlnum(c.text)

		test.Assert(t, "", c.exp, got, true)
	}
}
