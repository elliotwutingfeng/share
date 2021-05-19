// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type Response struct {
	Param        *Value
	FaultMessage string
	FaultCode    int32
	IsFault      bool
}

func (resp *Response) UnmarshalText(text []byte) (err error) {
	var (
		logp = "xmlrpc: Response"
		dec  = xml.NewDecoder(bytes.NewReader(text))
	)

	err = xmlBegin(dec)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = xmlMustStart(dec, elNameMethodResponse)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case elNameFault:
				err = resp.unmarshalFault(dec)
				if err != nil {
					return fmt.Errorf("%s: %w", logp, err)
				}
				found = true

			case elNameParams:
				resp.Param, err = xmlParseParam(dec, elNameParams)
				if err != nil {
					return fmt.Errorf("%s: %w", logp, err)
				}
				found = true

			default:
				return fmt.Errorf("%s: expecting <params> or <fault> got <%s>",
					logp, tok.Name.Local)
			}

		case xml.Comment, xml.CharData:
			token, err = dec.Token()
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

		default:
			return fmt.Errorf("%s: expecting <params> or <fault>, got token %T %+v",
				logp, token, tok)
		}
	}

	return nil
}

//
// unmarshalFault parse the XML fault error code and message.
//
func (resp *Response) unmarshalFault(dec *xml.Decoder) (err error) {
	resp.IsFault = true

	v, err := xmlParseValue(dec, elNameFault)
	if err != nil {
		return fmt.Errorf("unmarshalFault: %w", err)
	}

	resp.FaultCode = v.GetFieldAsInteger(memberNameFaultCode)
	resp.FaultMessage = v.GetFieldAsString(memberNameFaultString)

	return nil
}
