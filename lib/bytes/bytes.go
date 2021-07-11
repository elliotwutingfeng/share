// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bytes provide a library for working with byte or slice of bytes.
package bytes

import (
	"bytes"
	"fmt"
	"reflect"
	"unicode"
)

//
// AppendInt16 into slice of byte.
//
func AppendInt16(data *[]byte, v int16) {
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendInt32 into slice of byte.
//
func AppendInt32(data *[]byte, v int32) {
	*data = append(*data, byte(v>>24))
	*data = append(*data, byte(v>>16))
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendUint16 into slice of byte.
//
func AppendUint16(data *[]byte, v uint16) {
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendUint32 into slice of byte.
//
func AppendUint32(data *[]byte, v uint32) {
	*data = append(*data, byte(v>>24))
	*data = append(*data, byte(v>>16))
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// Concat merge one or more slice of byte or string in arguments into slice of
// byte.
// Any type that is not []byte or string in arguments will be ignored.
//
func Concat(sb []byte, args ...interface{}) (out []byte) {
	if len(sb) > 0 {
		out = append(out, sb...)
	}
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.String {
			out = append(out, arg.(string)...)
		}
		if v.Kind() != reflect.Slice {
			continue
		}
		if v.Len() == 0 {
			continue
		}
		if v.Index(0).Kind() != reflect.Uint8 {
			continue
		}
		b := v.Bytes()
		out = append(out, b...)
	}
	return
}

//
// Copy slice of bytes from parameter.
//
func Copy(src []byte) (dst []byte) {
	if len(src) == 0 {
		return
	}
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

//
// CutUntilToken cut line until we found token.
//
// If token found, it will return all cutted bytes before token, positition of
// byte after token, and boolean true.
//
// If no token found, it will return false.
//
// If `checkEsc` is true, token that is prefixed with escaped character
// '\' will be skipped.
//
//
func CutUntilToken(line, token []byte, startAt int, checkEsc bool) ([]byte, int, bool) {
	var (
		v              []byte
		p              int
		found, escaped bool
	)

	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return line, -1, false
	}
	if startAt < 0 {
		startAt = 0
	}

	for p = startAt; p < linelen; p++ {
		// Check if the escape character is used to escaped the
		// token ...
		if checkEsc && line[p] == '\\' {
			if escaped {
				// escaped already, its mean double '\\'
				v = append(v, '\\')
				escaped = false
			} else {
				escaped = true
			}
			continue
		}
		if line[p] != token[0] {
			if escaped {
				// ... turn out its not escaping token.
				v = append(v, '\\')
				escaped = false
			}
			v = append(v, line[p])
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found = IsTokenAt(line, token, p)

		// False alarm ...
		if !found {
			if escaped {
				// ... turn out its not escaping token.
				v = append(v, '\\')
				escaped = false
			}
			v = append(v, line[p])
			continue
		}

		// Found it, but if its prefixed with escaped char, then
		// we assumed it as non breaking token.
		if escaped {
			v = append(v, token...)
			p = p + tokenlen - 1
			escaped = false
			continue
		}

		// We found the token match in `line` at `p`
		return v, p + tokenlen, true
	}

	// We did not found it...
	return v, p, false
}

//
// EncloseRemove given a line, remove all bytes inside it, starting from
// `leftcap` until the `rightcap` and return cutted line and status to true.
//
// If no `leftcap` or `rightcap` is found, it will return line as is, and
// status will be false.
//
func EncloseRemove(line, leftcap, rightcap []byte) ([]byte, bool) {
	lidx := TokenFind(line, leftcap, 0)
	ridx := TokenFind(line, rightcap, lidx+1)

	if lidx < 0 || ridx < 0 || lidx >= ridx {
		return line, false
	}

	var newline []byte
	newline = append(newline, line[:lidx]...)
	newline = append(newline, line[ridx+len(rightcap):]...)
	newline, _ = EncloseRemove(newline, leftcap, rightcap)

	return newline, true
}

//
// EncloseToken will find `token` in `line` and enclose it with bytes from
// `leftcap` and `rightcap`.
// If at least one token found, it will return modified line with true status.
// If no token is found, it will return the same line with false status.
//
func EncloseToken(line, token, leftcap, rightcap []byte) (
	newline []byte,
	status bool,
) {
	enclosedLen := len(token)

	startat := 0
	for {
		foundat := TokenFind(line, token, startat)

		if foundat < 0 {
			newline = append(newline, line[startat:]...)
			break
		}

		newline = append(newline, line[startat:foundat]...)
		newline = append(newline, leftcap...)
		newline = append(newline, token...)
		newline = append(newline, rightcap...)

		startat = foundat + enclosedLen
	}
	if startat > 0 {
		status = true
	}

	return
}

//
// IsTokenAt return true if `line` at index `p` match with `token`,
// otherwise it will return false.
// Empty token always return false.
//
func IsTokenAt(line, token []byte, p int) bool {
	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return false
	}
	if p < 0 {
		p = 0
	}

	if p+tokenlen > linelen {
		return false
	}

	for x := 0; x < tokenlen; x++ {
		if line[p] != token[x] {
			return false
		}
		p++
	}
	return true
}

//
// PrintHex will print each byte in slice as hexadecimal value into N column
// length.
//
func PrintHex(title string, data []byte, col int) {
	var (
		start, x int
	)
	fmt.Print(title)
	for x = 0; x < len(data); x++ {
		if x%col == 0 {
			if x > 0 {
				fmt.Print(" ||")
			}
			for y := start; y < x; y++ {
				if data[y] >= 33 && data[y] <= 126 {
					fmt.Printf(" %c", data[y])
				} else {
					fmt.Print(" .")
				}
			}
			fmt.Printf("\n%4d -", x)
			start = x
		}

		fmt.Printf(" %02X", data[x])
	}
	rest := 16 - (x % col)
	if rest > 0 {
		for y := 0; y < rest; y++ {
			fmt.Print("   ")
		}
		fmt.Print(" ||")
	}
	for y := start; y < x; y++ {
		if data[y] >= 33 && data[y] <= 126 {
			fmt.Printf(" %c", data[y])
		} else {
			fmt.Print(" .")
		}
	}

	fmt.Println()
}

//
// ReadHexByte read two characters from data start from index "x" and convert
// them to byte.
//
func ReadHexByte(data []byte, x int) (b byte, ok bool) {
	if len(data) < x+2 {
		return 0, false
	}
	var y uint = 4
	for {
		switch {
		case data[x] >= '0' && data[x] <= '9':
			b |= (data[x] - '0') << y
		case data[x] >= 'A' && data[x] <= 'F':
			b |= (data[x] - ('A' - 10)) << y
		case data[x] >= 'a' && data[x] <= 'f':
			b |= (data[x] - ('a' - 10)) << y
		default:
			return b, false
		}
		if y == 0 {
			break
		}
		y -= 4
		x++
	}

	return b, true
}

//
// ReadInt16 will convert two bytes from data start at `x` into int16 and
// return it.
//
func ReadInt16(data []byte, x uint) int16 {
	return int16(data[x])<<8 | int16(data[x+1])
}

//
// ReadInt32 will convert four bytes from data start at `x` into int32 and
// return it.
//
func ReadInt32(data []byte, x uint) int32 {
	return int32(data[x])<<24 | int32(data[x+1])<<16 | int32(data[x+2])<<8 | int32(data[x+3])
}

//
// ReadUint16 will convert two bytes from data start at `x` into uint16 and
// return it.
//
func ReadUint16(data []byte, x uint) uint16 {
	return uint16(data[x])<<8 | uint16(data[x+1])
}

//
// ReadUint32 will convert four bytes from data start at `x` into uint32 and
// return it.
//
func ReadUint32(data []byte, x uint) uint32 {
	return uint32(data[x])<<24 | uint32(data[x+1])<<16 | uint32(data[x+2])<<8 | uint32(data[x+3])
}

//
// InReplace do a reverse replace on input, any characters that is not on
// allowed, will be replaced with character c.
//
func InReplace(in, allowed []byte, c byte) (out []byte) {
	if len(in) == 0 {
		return
	}

	out = make([]byte, len(in))
	copy(out, in)
	var found bool
	for x := 0; x < len(in); x++ {
		found = false
		for y := 0; y < len(allowed); y++ {
			if in[x] == allowed[y] {
				found = true
				break
			}
		}
		if !found {
			out[x] = c
		}
	}

	return
}

//
// Indexes returns the index of the all instance of token in s, or nil if
// token is not present in s.
//
func Indexes(s []byte, token []byte) (idxs []int) {
	if len(s) == 0 || len(token) == 0 {
		return nil
	}

	offset := 0
	for {
		idx := bytes.Index(s, token)
		if idx == -1 {
			break
		}
		idxs = append(idxs, offset+idx)
		skip := idx + len(token)
		offset += skip
		s = s[skip:]
	}
	return idxs
}

//
// MergeSpaces convert sequences of white spaces into single space ' '.
//
func MergeSpaces(in []byte) (out []byte) {
	var isSpace bool
	for _, c := range in {
		if c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == '\n' || c == '\r' {
			isSpace = true
			continue
		}
		if isSpace {
			out = append(out, ' ')
			isSpace = false
		}
		out = append(out, c)
	}
	if isSpace {
		out = append(out, ' ')
	}

	return out
}

//
// SkipAfterToken skip all bytes until matched token is found and return the
// index after the token and boolean true.
//
// If `checkEsc` is true, token that is prefixed with escaped character
// '\' will be considered as non-match token.
//
// If no token found it will return -1 and boolean false.
//
func SkipAfterToken(line, token []byte, startAt int, checkEsc bool) (int, bool) {
	linelen := len(line)
	escaped := false
	if startAt < 0 {
		startAt = 0
	}

	p := startAt
	for ; p < linelen; p++ {
		// Check if the escape character is used to escaped the
		// token.
		if checkEsc && line[p] == '\\' {
			escaped = true
			continue
		}
		if line[p] != token[0] {
			if escaped {
				escaped = false
			}
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found := IsTokenAt(line, token, p)

		// False alarm ...
		if !found {
			if escaped {
				escaped = false
			}
			continue
		}

		// Its matched, but if its prefixed with escaped char, then
		// we assumed it as non breaking token.
		if checkEsc && escaped {
			escaped = false
			continue
		}

		// We found the token at `p`
		p += len(token)
		return p, true
	}

	return p, false
}

//
// SnippetByIndexes take snippet in between of each index with minimum
// snippet length.  The sniplen is the length before and after index, not the
// length of all snippet.
//
func SnippetByIndexes(s []byte, indexes []int, sniplen int) (snippets [][]byte) {
	var start, end int
	for _, idx := range indexes {
		start = idx - sniplen
		if start < 0 {
			start = 0
		}
		end = idx + sniplen
		if end > len(s) {
			end = len(s)
		}

		snippets = append(snippets, s[start:end])
	}

	return snippets
}

//
// TokenFind return the first index of matched token in line, start at custom
// index.
// If "startat" parameter is less than 0, then it will be set to 0.
// If token is empty or no token found it will return -1.
//
func TokenFind(line, token []byte, startat int) (at int) {
	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return -1
	}
	if startat < 0 {
		startat = 0
	}

	y := 0
	at = -1
	for x := startat; x < linelen; x++ {
		if line[x] == token[y] {
			if y == 0 {
				at = x
			}
			y++
			if y == tokenlen {
				// we found it!
				return at
			}
		} else if at != -1 {
			// reset back
			y = 0
			at = -1
		}
	}
	// x run out before y
	if y < tokenlen {
		at = -1
	}

	return at
}

//
// WordIndexes returns the index of the all instance of word in s as long as
// word is separated by space or at the beginning or end of s.
//
func WordIndexes(s []byte, word []byte) (idxs []int) {
	tmp := Indexes(s, word)
	if len(tmp) == 0 {
		return nil
	}

	for _, idx := range tmp {
		x := idx - 1
		if x >= 0 {
			if !unicode.IsSpace(rune(s[x])) {
				continue
			}
		}
		x = idx + len(word)
		if x >= len(s) {
			idxs = append(idxs, idx)
			continue
		}
		if !unicode.IsSpace(rune(s[x])) {
			continue
		}
		idxs = append(idxs, idx)
	}

	return idxs
}

//
// WriteUint16 into slice of byte.
//
func WriteUint16(data *[]byte, x uint, v uint16) {
	(*data)[x] = byte(v >> 8)
	(*data)[x+1] = byte(v)
}

//
// WriteUint32 into slice of byte.
//
func WriteUint32(data *[]byte, x uint, v uint32) {
	(*data)[x] = byte(v >> 24)
	(*data)[x+1] = byte(v >> 16)
	(*data)[x+2] = byte(v >> 8)
	(*data)[x+3] = byte(v)
}
