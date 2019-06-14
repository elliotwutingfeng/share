//
// This benchmark is taken from github.com/gobwas/ws [1][2].
//
// [1] https://github.com/gobwas/ws/blob/master/server_test.go
// [2] https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb
//
package websocket

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	nonceKeySize = 16
	nonceSize    = 24 // base64.StdEncoding.EncodedLen(nonceKeySize)
)

type upgradeCase struct {
	label string
	nonce []byte
	req   *http.Request
}

//nolint:gochecknoglobals
var upgradeCases = []upgradeCase{
	{
		label: "base",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "lowercase",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			strings.ToLower(_hdrKeyUpgrade):    []string{"websocket"},
			strings.ToLower(_hdrKeyConnection): []string{"Upgrade"},
			strings.ToLower(_hdrKeyWSVersion):  []string{"13"},
		}),
	},
	{
		label: "uppercase",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"WEBSOCKET"},
			_hdrKeyConnection: []string{"UPGRADE"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "subproto",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
			_hdrKeyWSProtocol: []string{"a", "b", "c", "d"},
		}),
	},
	{
		label: "subproto_comma",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
			_hdrKeyWSProtocol: []string{"a, b, c, d"},
		}),
	},
	{
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:      []string{"websocket"},
			_hdrKeyConnection:   []string{"Upgrade"},
			_hdrKeyWSVersion:    []string{"13"},
			_hdrKeyWSExtensions: []string{"a;foo=1", "b;bar=2", "c", "d;baz=3"},
		}),
	},

	// Error cases.
	// ------------

	{
		label: "bad_http_method",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("POST", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_http_proto",
		nonce: mustMakeNonce(),
		req: setProto(1, 0, mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		})),
	},
	{
		label: "bad_host",
		nonce: mustMakeNonce(),
		req: withoutHeader("Host", mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		})),
	},
	{
		label: "bad_upgrade",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_upgrade",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			"X-Custom-Header": []string{"value"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_upgrade",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"not-websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_connection",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:   []string{"websocket"},
			_hdrKeyWSVersion: []string{"13"},
		}),
	},
	{
		label: "bad_connection",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"not-upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_sec_version_x",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
		}),
	},
	{
		label: "bad_sec_version",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"upgrade"},
			_hdrKeyWSVersion:  []string{"15"},
		}),
	},
	{
		label: "bad_sec_key",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
	{
		label: "bad_sec_key",
		nonce: mustMakeNonce(),
		req: mustMakeRequest("GET", http.Header{
			_hdrKeyUpgrade:    []string{"websocket"},
			_hdrKeyConnection: []string{"Upgrade"},
			_hdrKeyWSVersion:  []string{"13"},
		}),
	},
}

func BenchmarkUpgrader(b *testing.B) {
	for _, bench := range upgradeCases {
		bench.req.Header.Set(_hdrKeyWSKey, string(bench.nonce))

		u := Server{}

		reqBytes := dumpRequest(bench.req)

		b.Run(bench.label, func(b *testing.B) {
			conn := make([][]byte, b.N)
			for i := 0; i < b.N; i++ {
				conn[i] = reqBytes
			}

			i := new(int64)

			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c := conn[atomic.AddInt64(i, 1)-1]
					_, _, _ = u.handleUpgrade(c)
				}
			})
		})
	}
}

func mustMakeRequest(method string, headers http.Header) *http.Request {
	req, err := http.NewRequest(method, "ws://example.org", nil)
	if err != nil {
		panic(err)
	}
	req.Header = headers
	return req
}

func setProto(major, minor int, req *http.Request) *http.Request {
	req.ProtoMajor = major
	req.ProtoMinor = minor
	return req
}

func withoutHeader(header string, req *http.Request) *http.Request {
	if strings.EqualFold(header, "Host") {
		req.URL.Host = ""
		req.Host = ""
	} else {
		delete(req.Header, header)
	}
	return req
}

// initNonce fills given slice with random base64-encoded nonce bytes.
func initNonce(dst []byte) {
	// NOTE: bts does not escapes.
	bts := make([]byte, nonceKeySize)
	if _, err := rand.Read(bts); err != nil { //nolint:gas
		panic(fmt.Sprintf("rand read error: %s", err))
	}
	base64.StdEncoding.Encode(dst, bts)
}

func mustMakeNonce() (ret []byte) {
	ret = make([]byte, nonceSize)
	initNonce(ret)
	return
}

func dumpRequest(req *http.Request) []byte {
	bts, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	return bts
}
