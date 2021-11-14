// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// SectionQuestion The question section is used to carry the "question" in
// most queries, i.e., the parameters that define what is being asked.  The
// section contains QDCOUNT (usually 1) entries, each of the following format:
//
type SectionQuestion struct {
	// A domain name represented as a sequence of labels, where each label
	// consists of a length octet followed by that number of octets.  The
	// domain name terminates with the zero length octet for the null
	// label of the root.  Note that this field may be an odd number of
	// octets; no padding is used.
	Name string

	// A two octet code which specifies the type of the query.  The values
	// for this field include all codes valid for a TYPE field, together
	// with some more general codes which can match more than one type of
	// RR.
	Type RecordType

	// A two octet code that specifies the class of the query.  For
	// example, the QCLASS field is IN for the Internet.
	Class RecordClass
}

//
// Reset the message question field to it's default values for query.
//
func (question *SectionQuestion) Reset() {
	question.Name = question.Name[:0]
	question.Type = RecordTypeA
	question.Class = RecordClassIN
}

//
// size return the section question size, length of name + 2 (1 octet for
// beginning size plus 1 octet for end of label) + 2 octets of
// rtype + 2 octets of rclass
//
func (question *SectionQuestion) size() int {
	return len(question.Name) + 6
}

//
// String will return the string representation of section question structure.
//
func (question *SectionQuestion) String() string {
	return fmt.Sprintf("&{Name:%s Type:%s}", question.Name,
		RecordTypeNames[question.Type])
}

//
// unpack the DNS question section.
//
func (question *SectionQuestion) unpack(packet []byte) (err error) {
	if len(packet) == 0 {
		return nil
	}

	var sb strings.Builder
	count := packet[0]
	x := 1

	for {
		if count == 0 {
			sb.WriteByte('.')
			count = packet[x]
			x++
			if x >= len(packet) {
				return fmt.Errorf("SectionQuestion.unpack: invalid question %q", packet)
			}
			continue
		}
		for y := byte(0); y < count; y++ {
			if packet[x] >= 'A' && packet[x] <= 'Z' {
				packet[x] += 32
			}
			sb.WriteByte(packet[x])
			x++
			if x >= len(packet) {
				return fmt.Errorf("SectionQuestion.unpack: invalid question %q", packet)
			}
		}
		count = packet[x]
		x++
		if count == 0 {
			break
		}
		if x >= len(packet) {
			return fmt.Errorf("SectionQuestion.unpack: invalid question %q", packet)
		}
		sb.WriteByte('.')
	}

	if x+4 > len(packet) {
		return fmt.Errorf("SectionQuestion.unpack: invalid question %q", packet)
	}

	question.Name = sb.String()
	question.Type = RecordType(libbytes.ReadUint16(packet, uint(x)))
	x += 2
	question.Class = RecordClass(libbytes.ReadUint16(packet, uint(x)))

	return nil
}
