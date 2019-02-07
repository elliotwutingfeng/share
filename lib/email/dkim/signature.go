// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	sepHeaders        = []byte{':'}
	sepMethods        = []byte{':'}
	sepPresentHeaders = []byte{'|'}
)

//
// Signature represents the value of DKIM-Signature header field tag.
//
type Signature struct {
	// Version of specification.
	// It MUST have the value "1" for compliant with RFC 6376.
	// ("v=", text, REQUIRED).
	Version []byte

	// Algorithm used to generate the signature.
	// Valid values is "rsa-sha1" or "rsa-sha256".
	// ("a=", text, REQUIRED).
	Alg *SignAlg

	// Signer domain
	// ("d=", text, REQUIRED).
	SDID []byte

	// The selector subdividing the namespace for the "d=" tag.
	// ("s=", text, REQUIRED).
	Selector []byte

	// List of header field names included in signing or verifying.
	// ("h=", text, REQUIRED).
	Headers [][]byte

	// The hash of canonicalized body data.
	// ("bh=", base64, REQUIRED).
	BodyHash []byte

	// The signature data.
	// ("b=", base64, REQUIRED)
	Value []byte

	// RECOMMENDED fields

	// Time when signature created, in UNIX timestamp.
	// ("t=", text, RECOMMENDED).
	CreatedAt uint64

	// Expiration time, in UNIX timestamp.
	// ("x=", text, RECOMMENDED).
	ExpiredAt uint64

	// OPTIONAL fields

	// Type of canonicalization for header.  Default is "simple".
	// ("c=header/body", text, OPTIONAL).
	CanonHeader *Canon

	// Type of canonicalization for body.  Default is "simple".
	// ("c=header/body", text, OPTIONAL).
	CanonBody *Canon

	// List of header field name and value that present when the message
	// is signed.
	// ("z=", dkim-quoted-printable, OPTIONAL).  Default is null.
	PresentHeaders [][]byte

	// The agent or user identifier.
	// Default is "@" + "d=" value)
	// ("i=", dkim-quoted-printable, OPTIONAL).
	AUID []byte

	// The number of octets in body, after canonicalization, included when
	// computing hash.
	// If nil, its means entire body is signed.
	// If 0, its means the body is not signed.
	// ("l=", text, OPTIONAL).
	BodyLength *uint64

	// QMethod define a type and option used to retrieve the public keys.
	// ("q=type/option", text, OPTIONAL).  Default is "dns/txt".
	QMethod *QueryMethod

	// raw contains original Signature field value, for Simple
	// canonicalization.
	raw []byte
}

//
// Parse DKIM-Signature field value.
// The signature value MUST be end with CRLF.
//
func Parse(value []byte) (sig *Signature, err error) {
	if len(value) == 0 {
		return nil, nil
	}

	l := len(value)
	if value[l-2] != '\r' && value[l-1] != '\n' {
		return nil, errors.New("dkim: value must end with CRLF")
	}

	p := newParser(value)

	sig = &Signature{}
	for {
		tag, err := p.fetchTag()
		if err != nil {
			return nil, err
		}
		if tag == nil {
			break
		}
		if tag.key == tagUnknown {
			continue
		}
		err = sig.set(tag)
		if err != nil {
			return nil, err
		}
	}

	sig.raw = value

	return sig, err
}

//
// Relaxed return the relaxed canonicalization of Signature ordered by tag
// priority: required, recommended, and optional.
// Recommended and optional field values will be printed only if its not
// empty.
//
func (sig *Signature) Relaxed() []byte {
	var bb bytes.Buffer

	_, _ = fmt.Fprintf(&bb, "v=%s; a=%s; d=%s; s=%s;\r\n\t"+
		"h=%s;\r\n\tbh=%s;\r\n\tb=%s;\r\n\t",
		sig.Version, signAlgNames[*sig.Alg], sig.SDID, sig.Selector,
		bytes.Join(sig.Headers, sepHeaders), sig.BodyHash, sig.Value)

	if sig.CreatedAt > 0 {
		_, _ = fmt.Fprintf(&bb, "t=%d; ", sig.CreatedAt)
	}
	if sig.ExpiredAt > 0 {
		_, _ = fmt.Fprintf(&bb, "x=%d; ", sig.ExpiredAt)
	}

	if sig.CanonHeader != nil {
		_, _ = fmt.Fprintf(&bb, "c=%s", canonNames[*sig.CanonHeader])

		if sig.CanonBody != nil {
			_, _ = fmt.Fprintf(&bb, "/%s;\r\n\t",
				canonNames[*sig.CanonBody])
		} else {
			bb.WriteString(";\r\n\t")
		}
	}

	if len(sig.PresentHeaders) > 0 {
		_, _ = fmt.Fprintf(&bb, "z=%s;\r\n\t",
			bytes.Join(sig.PresentHeaders, []byte{'|', '\r', '\n', '\t', ' '}))
	}

	if len(sig.AUID) > 0 {
		_, _ = fmt.Fprintf(&bb, "i=%s; ", sig.AUID)
	}
	if sig.BodyLength != nil {
		_, _ = fmt.Fprintf(&bb, "l=%d; ", *sig.BodyLength)
	}
	if sig.QMethod != nil {
		_, _ = fmt.Fprintf(&bb, "q=%s/%s;\r\n",
			queryTypeNames[sig.QMethod.Type],
			queryOptionNames[sig.QMethod.Option])
	} else {
		bb.WriteString("\r\n")
	}

	return bb.Bytes()
}

//
// Simple return the "simple" canonicalization of Signature.
//
func (sig *Signature) Simple() []byte {
	return sig.raw
}

//
// Verify the tag list.
//
// Rules of tags,
//
// *  Tags MUST NOT duplicate.  This was already handled by parser.
//
// *  All the required fields MUST NOT be empty or nil.
//
// *  The "h=" tag MUST include the "From" header field
//
// *  The value of the "x=" tag MUST be greater than the value of the "t=" tag
// if both are present.
//
// *  The "d=" value MUST be the same or parent domain of "i="
//
func (sig *Signature) Verify() (err error) {
	if len(sig.Version) == 0 || sig.Version[0] != '1' {
		return fmt.Errorf("dkim: invalid version number: '%s'", sig.Version)
	}
	if sig.Alg == nil {
		return fmt.Errorf("dkim: tag algorithm 'a=' is not defined")
	}
	if len(sig.SDID) == 0 {
		return fmt.Errorf("dkim: tag SDID 'd=' is not defined")
	}
	if len(sig.Selector) == 0 {
		return fmt.Errorf("dkim: tag selector 's=' is not defined")
	}
	if len(sig.Headers) == 0 {
		return fmt.Errorf("dkim: tag header 'h=' is empty")
	}

	found := false
	for x := 0; x < len(sig.Headers); x++ {
		fmt.Printf("h[%d]=%s\n", x, sig.Headers[x])
		if bytes.EqualFold(sig.Headers[x], []byte("from")) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dkim: 'From' field not signed")
	}

	if len(sig.BodyHash) == 0 {
		return fmt.Errorf("dkim: tag body hash 'bh=' is empty")
	}
	if len(sig.Value) == 0 {
		return fmt.Errorf("dkim: tag signature 'h=' is empty")
	}

	if sig.ExpiredAt != 0 {
		if sig.CreatedAt != 0 && sig.ExpiredAt < sig.CreatedAt {
			return fmt.Errorf("dkim: invalid expiration/creation time")
		}
		exp := time.Unix(int64(sig.ExpiredAt), 0)
		now := time.Now().Add(time.Hour * -1).Unix()
		if uint64(now) > sig.ExpiredAt {
			return fmt.Errorf("dkim: signature is expired at '%s'", exp)
		}
	}
	if len(sig.AUID) > 0 {
		bb := bytes.Split(sig.AUID, []byte{'@'})
		if len(bb) != 2 {
			return fmt.Errorf("dkim: missing AUID domain: '%s'", sig.AUID)
		}
		fmt.Printf("AUID domain: %s\n", bb[1])
		if !bytes.HasSuffix(bb[1], sig.SDID) {
			return fmt.Errorf("dkim: invalid AUID: '%s'", sig.AUID)
		}
	}

	return nil
}

//
// set the signature field value with value from tag.
//
func (sig *Signature) set(t *tag) (err error) {
	if t == nil {
		return
	}

	var l uint64

	switch t.key {
	case tagVersion:
		sig.Version = t.value

	case tagAlg:
		for k, name := range signAlgNames {
			if bytes.Equal(t.value, name) {
				sig.Alg = &k
				return nil
			}
		}
		err = fmt.Errorf("dkim: unknown algorithm: '%s'", t.value)

	case tagSDID:
		sig.SDID = t.value

	case tagHeaders:
		sig.Headers = bytes.Split(t.value, sepHeaders)

	case tagSelector:
		sig.Selector = t.value

	case tagBodyHash:
		sig.BodyHash = t.value

	case tagSignature:
		sig.Value = t.value

	case tagCreatedAt:
		sig.CreatedAt, err = strconv.ParseUint(string(t.value), 10, 64)

	case tagExpiredAt:
		sig.ExpiredAt, err = strconv.ParseUint(string(t.value), 10, 64)

	case tagCanon:
		err = sig.setCanons(t.value)

	case tagPresentHeaders:
		sig.PresentHeaders = bytes.Split(t.value, sepPresentHeaders)

	case tagAUID:
		sig.AUID = t.value

	case tagBodyLength:
		l, err = strconv.ParseUint(string(t.value), 10, 64)
		if err == nil {
			sig.BodyLength = &l
		}

	case tagQueryMethod:
		sig.setQueryMethods(t.value)
	}

	return err
}

//
// setCanons set Signature canonicalization algorithm for header and body
// based on text in "v".
//
func (sig *Signature) setCanons(v []byte) (err error) {
	var canonHeader, canonBody []byte

	canons := bytes.Split(v, []byte{'/'})

	switch len(canons) {
	case 0:
	case 1:
		canonHeader = canons[0]
	case 2:
		canonHeader = canons[0]
		canonBody = canons[1]
	default:
		err = fmt.Errorf("dkim: invalid canonicalization: '%s'", v)
	}

	t, err := parseCanonValue(canonHeader)
	if err != nil {
		return err
	}
	if t != nil {
		sig.CanonHeader = t

		t, err = parseCanonValue(canonBody)
		if err != nil {
			return err
		}
		if t != nil {
			sig.CanonBody = t
		}
	}

	return nil
}

//
// parseCanonValue parse canonicalization name and return their numeric type.
//
func parseCanonValue(v []byte) (*Canon, error) {
	if len(v) == 0 {
		return nil, nil
	}
	for k, cname := range canonNames {
		if bytes.Equal(v, cname) {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("dkim: invalid canonicalization: '%s'", v)
}

//
// setQueryMethods parse list of query methods and set Signature.QueryMethod
// based on first match.
//
func (sig *Signature) setQueryMethods(v []byte) {
	methods := bytes.Split(v, sepMethods)

	for _, m := range methods {
		var qtype, qopt []byte

		kv := bytes.Split(m, []byte{'/'})
		switch len(kv) {
		case 0:
		case 1:
			qtype = kv[0]
		case 2:
			qtype = kv[0]
			qopt = kv[1]
		}
		err := sig.setQueryMethod(qtype, qopt)
		if err != nil {
			sig.QMethod = nil
			// Ignore error, use default query method.
		}
	}
}

//
// setQueryMethod set Signature query type and option.
//
func (sig *Signature) setQueryMethod(qtype, qopt []byte) (err error) {
	if len(qtype) == 0 {
		return nil
	}

	sig.QMethod = &QueryMethod{}

	found := false
	for k, typ := range queryTypeNames {
		if bytes.Equal(qtype, typ) {
			sig.QMethod.Type = k
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dkim: unknown query type: '%s'", qtype)
	}
	if len(qopt) == 0 {
		return nil
	}

	found = false
	for k, opt := range queryOptionNames {
		if bytes.Equal(qopt, opt) {
			sig.QMethod.Option = k
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dkim: unknown query option: '%s'", qopt)
	}

	return nil
}