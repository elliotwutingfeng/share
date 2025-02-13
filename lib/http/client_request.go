// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClientRequest define the parameters for each Client methods.
type ClientRequest struct {
	// Headers additional header to be send on request.
	// This field is optional.
	Headers http.Header

	//
	// Params define parameter to be send on request.
	// This field is optional.
	//
	// For Method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the
	// params value should be nil or url.Values.
	// If its url.Values, then the params will be encoded as query
	// parameters.
	//
	// For Method PATCH, POST, or PUT; the Params will converted based on
	// Type rules below,
	//
	// * If Type is RequestTypeQuery and Params is url.Values it will be
	// added as query parameters in the Path.
	//
	// * If Type is RequestTypeForm and Params is url.Values it will be
	// added as URL encoded in the body.
	//
	// * If Type is RequestTypeMultipartForm and Params is
	// map[string][]byte, then it will be converted as multipart form in
	// the body.
	//
	// * If Type is RequestTypeJSON and Params is not nil, the params will
	// be encoded as JSON in body using json.Encode().
	//
	Params interface{}

	// The Path to resource on the server.
	// This field is required, if its empty default to "/".
	Path string

	// The HTTP method of request.
	// This field is optional, if its empty default to RequestMethodGet
	// (GET).
	Method RequestMethod

	// The Type of request.
	// This field is optional, it's affect how the Params field encoded in
	// the path or body.
	Type RequestType
}

// toHttpRequest convert the ClientRequest into the standard http.Request.
func (creq *ClientRequest) toHttpRequest(client *Client) (httpReq *http.Request, err error) {
	var (
		logp              = "toHttpRequest"
		paramsAsUrlValues url.Values
		paramsAsJSON      []byte
		contentType       = creq.Type.String()
		path              strings.Builder
		body              io.Reader
		strBody           string
		isParamsUrlValues bool
	)

	if client != nil {
		path.WriteString(client.opts.ServerUrl)
	}
	path.WriteString(creq.Path)
	paramsAsUrlValues, isParamsUrlValues = creq.Params.(url.Values)

	switch creq.Method {
	case RequestMethodGet,
		RequestMethodConnect,
		RequestMethodDelete,
		RequestMethodHead,
		RequestMethodOptions,
		RequestMethodTrace:

		if isParamsUrlValues {
			path.WriteString("?")
			path.WriteString(paramsAsUrlValues.Encode())
		}

	case RequestMethodPatch,
		RequestMethodPost,
		RequestMethodPut:
		switch creq.Type {
		case RequestTypeQuery:
			if isParamsUrlValues {
				path.WriteString("?")
				path.WriteString(paramsAsUrlValues.Encode())
			}

		case RequestTypeForm:
			if isParamsUrlValues {
				strBody = paramsAsUrlValues.Encode()
				body = strings.NewReader(strBody)
			}

		case RequestTypeMultipartForm:
			paramsAsMultipart, ok := creq.Params.(map[string][]byte)
			if ok {
				contentType, strBody, err = generateFormData(paramsAsMultipart)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}
				body = strings.NewReader(strBody)
			}

		case RequestTypeJSON:
			if creq.Params != nil {
				paramsAsJSON, err = json.Marshal(creq.Params)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}
				body = bytes.NewReader(paramsAsJSON)
			}
		}
	}

	httpReq, err = http.NewRequest(creq.Method.String(), path.String(), body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if client != nil {
		setHeaders(httpReq, client.opts.Headers)
	}
	setHeaders(httpReq, creq.Headers)

	if len(contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, contentType)
	}

	return httpReq, nil
}
