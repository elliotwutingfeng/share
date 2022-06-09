// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

const (
	_maxQueue = 128

	_resUpgradeOK = "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Accept: "

	_resStatusOK = "HTTP/1.1 200 OK\r\n" +
		"Content-Type: %s\r\n" +
		"Content-Length: %d\r\n" +
		"\r\n" +
		"%s"
)

// Server for websocket.
type Server struct {
	poll libnet.Poll

	Clients *ClientManager

	// Options for server, set by calling NewServer.
	// This field is exported only for reference, for example logging in
	// the Options when server started.
	// Modifying the value of Options after server has been started may
	// cause undefined effects.
	Options *ServerOptions

	chUpgrade chan int
	running   chan struct{}

	routes *rootRoute

	// handlePong callback that will be called after receiving control
	// PONG frame from client. Default is nil, used only for testing.
	handlePong HandlerFrameFn

	sock int

	allowRsv1 bool
	allowRsv2 bool
	allowRsv3 bool
}

// NewServer will create new web-socket server that listen on specific port
// number.
func NewServer(opts *ServerOptions) (serv *Server) {
	if opts == nil {
		opts = &ServerOptions{}
	}

	serv = &Server{
		Options: opts,
		Clients: newClientManager(),
		routes:  newRootRoute(),
		running: make(chan struct{}, 1),
	}

	opts.init()

	if opts.HandleBin == nil {
		opts.HandleBin = serv.handleBin
	}
	if opts.HandleText == nil {
		opts.HandleText = serv.handleText
	}

	return serv
}

// AllowReservedBits allow receiving frame with RSV1, RSV2, or RSV3 bit set.
// Calling this function means server has negotiated the extension that use
// the reserved bits through handshake with client using HandleAuth.
//
// If a nonzero value is received in reserved bits and none of the negotiated
// extensions defines the meaning of such a nonzero value, server will close
// the connection (RFC 6455, section 5.2).
func (serv *Server) AllowReservedBits(one, two, three bool) {
	serv.allowRsv1 = one
	serv.allowRsv2 = two
	serv.allowRsv3 = three
}

func (serv *Server) createSockServer() (err error) {
	serv.sock, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return
	}

	err = unix.SetsockoptInt(serv.sock, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		return
	}

	var (
		host    string
		strPort string
		addr    unix.SockaddrInet4
	)

	host, strPort, err = net.SplitHostPort(serv.Options.Address)
	if err != nil {
		return
	}

	addr.Port, err = strconv.Atoi(strPort)
	if err != nil {
		return
	}

	copy(addr.Addr[:], net.ParseIP(host).To4())

	err = unix.Bind(serv.sock, &addr)
	if err != nil {
		return
	}

	err = unix.Listen(serv.sock, _maxQueue)

	return
}

// RegisterTextHandler register specific function to be called by server when
// request opcode is text, and method and target matched with Request.
func (serv *Server) RegisterTextHandler(method, target string, handler RouteHandler) (err error) {
	var logp string = "RegisterTextHandler"

	if len(method) == 0 {
		return fmt.Errorf("%s: empty method", logp)
	}
	if len(target) == 0 {
		return fmt.Errorf("%s: empty target", logp)
	}
	if handler == nil {
		return fmt.Errorf("%s: empty handler", logp)
	}

	err = serv.routes.add(method, target, handler)
	if err != nil {
		return fmt.Errorf("%s: %s %s: %w", logp, method, target, err)
	}

	return nil
}

func (serv *Server) handleError(conn int, code int, msg string) {
	var (
		rspBody = "HTTP/1.1 " + strconv.Itoa(code) + " " + msg + "\r\n\r\n"

		err error
	)

	err = Send(conn, []byte(rspBody))
	if err != nil {
		log.Println("websocket: server.handleError: " + err.Error())
	}

	unix.Close(conn)
}

// handleUpgrade parse and validate websocket HTTP handshake from client.
// If HandleAuth is not nil, the HTTP handshake will be passed to that
// function to allow custom authentication.
//
// On success it will return the context from authentication and the WebSocket
// key.
func (serv *Server) handleUpgrade(hs *Handshake) (
	ctx context.Context, key []byte, err error,
) {
	err = hs.parse()
	if err == nil {
		key = libbytes.Copy(hs.Key)
		if serv.Options.HandleAuth != nil {
			ctx, err = serv.Options.HandleAuth(hs)
		}
	}

	hs.reset(nil)
	_handshakePool.Put(hs)

	return ctx, key, err
}

// clientAdd add the new client connection to epoll and to list of clients.
func (serv *Server) clientAdd(ctx context.Context, conn int) (err error) {
	err = serv.poll.RegisterRead(conn)
	if err != nil {
		return fmt.Errorf("websocket: Server.clientAdd: " + err.Error())
	}

	if ctx != nil {
		serv.Clients.add(ctx, conn)

		if serv.Options.HandleClientAdd != nil {
			go serv.Options.HandleClientAdd(ctx, conn)
		}
	}

	return nil
}

// ClientRemove remove client connection from server.
func (serv *Server) ClientRemove(conn int) {
	var (
		ctx context.Context
		err error
	)

	ctx, _ = serv.Clients.Context(conn)

	if ctx != nil && serv.Options.HandleClientRemove != nil {
		serv.Options.HandleClientRemove(ctx, conn)
	}

	serv.Clients.remove(conn)

	err = serv.poll.UnregisterRead(conn)
	if err != nil {
		log.Println("websocket: server.ClientRemove: " + err.Error())
	}

	err = unix.Close(conn)
	if err != nil {
		log.Println("websocket: server.ClientRemove: " + err.Error())
	}
}

func (serv *Server) upgrader() {
	var (
		ctx      context.Context
		hs       *Handshake
		httpRes  string
		wsAccept string
		key      []byte
		packet   []byte
		conn     int
		err      error
	)

	for conn = range serv.chUpgrade {
		packet, err = Recv(conn)
		if err != nil {
			log.Println("websocket: server.upgrader: " + err.Error())
			unix.Close(conn)
			continue
		}
		if len(packet) == 0 {
			unix.Close(conn)
			continue
		}

		hs, err = newHandshake(packet)
		if err != nil {
			serv.handleError(conn, http.StatusBadRequest, err.Error())
			continue
		}

		if hs.URL.Path == serv.Options.StatusPath {
			serv.handleStatus(conn)
			continue
		}
		if hs.URL.Path != serv.Options.ConnectPath {
			serv.handleError(conn, http.StatusNotFound, "unknown path")
			continue
		}

		ctx, key, err = serv.handleUpgrade(hs)
		if err != nil {
			serv.handleError(conn, http.StatusBadRequest, err.Error())
			continue
		}

		wsAccept = generateHandshakeAccept(key)

		httpRes = _resUpgradeOK + wsAccept + "\r\n\r\n"

		err = Send(conn, []byte(httpRes))
		if err != nil {
			log.Println("websocket: server.upgrader: Send: ",
				err.Error())
			unix.Close(conn)
			continue
		}

		if ctx == nil {
			ctx = context.Background()
		}

		err = serv.clientAdd(ctx, conn)
		if err != nil {
			log.Println("websocket: server.upgrader: clientAdd: ",
				err.Error())
			unix.Close(conn)
		}
	}
}

// handleFragment will handle continuation frame (fragmentation).
//
// (RFC 6455 Section 5.4 Page 34)
// A fragmented message consists of a single frame with the FIN bit
// clear and an opcode other than 0, followed by zero or more frames
// with the FIN bit clear and the opcode set to 0, and terminated by
// a single frame with the FIN bit set and an opcode of 0.  A
// fragmented message is conceptually equivalent to a single larger
// message whose payload is equal to the concatenation of the
// payloads of the fragments in order; however, in the presence of
// extensions, this may not hold true as the extension defines the
// interpretation of the "Extension data" present.  For instance,
// "Extension data" may only be present at the beginning of the first
// fragment and apply to subsequent fragments, or there may be
// "Extension data" present in each of the fragments that applies
// only to that particular fragment.  In the absence of "Extension
// data", the following example demonstrates how fragmentation works.
//
// EXAMPLE: For a text message sent as three fragments, the first
// fragment would have an opcode of 0x1 and a FIN bit clear, the
// second fragment would have an opcode of 0x0 and a FIN bit clear,
// and the third fragment would have an opcode of 0x0 and a FIN bit
// that is set.
// (RFC 6455 Section 5.4 Page 34)
func (serv *Server) handleFragment(conn int, req *Frame) (isInvalid bool) {
	var (
		frames *Frames
		frame  *Frame
		ok     bool
	)

	frames, ok = serv.Clients.getFrames(conn)

	if debug.Value >= 3 {
		fmt.Printf("websocket: Server.handleFragment: frame: {fin:%d opcode:%d masked:%d len:%d, payload.len:%d}\n",
			req.fin, req.opcode, req.masked, req.len,
			len(req.payload))
	}

	if frames == nil {
		if req.opcode == OpcodeCont {
			// If a connection does not have continuous frame,
			// then current frame opcode must not be 0.
			serv.handleBadRequest(conn)
			return true
		}
	} else if req.opcode != OpcodeCont {
		// If a connection have continuous frame, the next frame
		// opcode must be 0.
		serv.handleBadRequest(conn)
		return true
	}

	if req.fin == 0 {
		if uint64(len(req.payload)) < req.len {
			// Continuous frame with unfinished payload.
			serv.Clients.setFrame(conn, req)
		} else {
			if frames == nil {
				frames = new(Frames)
			}
			frames.Append(req)
			serv.Clients.setFrame(conn, nil)
			serv.Clients.setFrames(conn, frames)
		}
		return false
	}

	if !ok && uint64(len(req.payload)) < req.len {
		// Final frame with unfinished payload.
		serv.Clients.setFrame(conn, req)
		return false
	}

	frame = serv.Clients.finFrames(conn, req)

	if frame.opcode == OpcodeText {
		if !utf8.Valid(frame.payload) {
			serv.handleInvalidData(conn)
			return true
		}
		go serv.Options.HandleText(conn, frame.payload)
	} else {
		go serv.Options.HandleBin(conn, frame.payload)
	}

	return false
}

// handleFrame handle a single frame from client.
func (serv *Server) handleFrame(conn int, frame *Frame) (isClosing bool) {
	if !frame.isValid(true, serv.allowRsv1, serv.allowRsv2, serv.allowRsv3) {
		serv.handleBadRequest(conn)
		return true
	}

	switch frame.opcode {
	case OpcodeCont, OpcodeText, OpcodeBin:
		var isInvalid bool = serv.handleFragment(conn, frame)
		if isInvalid {
			isClosing = true
		}
	case OpcodeDataRsv3, OpcodeDataRsv4, OpcodeDataRsv5, OpcodeDataRsv6,
		OpcodeDataRsv7:
		serv.handleBadRequest(conn)
		isClosing = true
	case OpcodeClose:
		serv.handleClose(conn, frame)
		isClosing = true
	case OpcodePing:
		serv.handlePing(conn, frame)
	case OpcodePong:
		if serv.handlePong != nil {
			go serv.handlePong(conn, frame)
		}
	case OpcodeControlRsvB, OpcodeControlRsvC, OpcodeControlRsvD,
		OpcodeControlRsvE, OpcodeControlRsvF:
		if serv.Options.HandleRsvControl != nil {
			serv.Options.HandleRsvControl(conn, frame)
		} else {
			serv.handleClose(conn, frame)
			isClosing = true
		}
	}
	return isClosing
}

// handleText message from client.
func (serv *Server) handleText(conn int, payload []byte) {
	var (
		handler RouteHandler
		err     error
		ctx     context.Context
		req     *Request
		res     *Response
		ok      bool
	)

	res = _resPool.Get().(*Response)
	res.reset()

	ctx, ok = serv.Clients.Context(conn)
	if !ok {
		err = errors.New("client context not found")
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		goto out
	}

	req = _reqPool.Get().(*Request)
	req.reset()

	err = json.Unmarshal(payload, req)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		goto out
	}

	handler, err = req.unpack(serv.routes)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = req.Target
		goto out
	}
	if handler == nil {
		res.Code = http.StatusNotFound
		res.Message = req.Method + " " + req.Target
		goto out
	}

	req.Conn = conn

	*res = handler(ctx, req)

out:
	if req != nil {
		res.ID = req.ID
		_reqPool.Put(req)
	}

	err = serv.sendResponse(conn, res)
	if err != nil {
		serv.ClientRemove(conn)
	}

	_resPool.Put(res)
}

// handleBin message from client.  This is the dummy handler, that can be
// overwritten by implementer.
func (serv *Server) handleBin(conn int, payload []byte) {}

func (serv *Server) handleStatus(conn int) {
	var (
		contentType string
		data        []byte
	)

	if serv.Options.HandleStatus == nil {
		contentType = "text/plain"
		data = []byte("OK")
	} else {
		contentType, data = serv.Options.HandleStatus()
	}

	var (
		res string = fmt.Sprintf(_resStatusOK, contentType, len(data), data)
		err error
	)

	err = Send(conn, []byte(res))
	if err != nil {
		log.Println("websocket /health: Send: ", err.Error())
	}

	unix.Close(conn)
}

// handleClose request from client.
func (serv *Server) handleClose(conn int, req *Frame) {
	switch {
	case req.closeCode == 0:
		req.closeCode = StatusBadRequest
	case req.closeCode < StatusNormal:
		req.closeCode = StatusBadRequest
	case req.closeCode == 1004:
		// Reserved.  The specific meaning might be defined in the future.
		req.closeCode = StatusBadRequest
	case req.closeCode == 1005:
		// 1005 is a reserved value and MUST NOT be set as a status
		// code in a Close control frame by an endpoint.  It is
		// designated for use in applications expecting a status code
		// to indicate that no status code was actually present.
		req.closeCode = StatusBadRequest
	case req.closeCode == 1006:
		//  1006 is a reserved value and MUST NOT be set as a status
		//  code in a Close control frame by an endpoint.  It is
		//  designated for use in applications expecting a status code
		//  to indicate that the connection was closed abnormally,
		//  e.g., without sending or receiving a Close control frame.
		req.closeCode = StatusBadRequest
	case req.closeCode >= 1015 && req.closeCode <= 2999:
		req.closeCode = StatusBadRequest
	case req.closeCode >= 3000 && req.closeCode <= 3999:
		// Status codes in the range 3000-3999 are reserved for use by
		// libraries, frameworks, and applications.  These status
		// codes are registered directly with IANA.  The
		// interpretation of these codes is undefined by this
		// protocol.
	case req.closeCode >= 4000 && req.closeCode <= 4999:
		// Status codes in the range 4000-4999 are reserved for
		// private use and thus can't be registered.  Such codes can
		// be used by prior agreements between WebSocket applications.
		// The interpretation of these codes is undefined by this
		// protocol.
	}
	if len(req.payload) >= 2 {
		// Cut the close code from actual payload.
		req.payload = req.payload[2:]
		if !utf8.Valid(req.payload) {
			req.closeCode = StatusBadRequest
		}
	}

	var (
		packet []byte = NewFrameClose(false, req.closeCode, req.payload)
		err    error
	)

	if debug.Value >= 3 {
		fmt.Printf("websocket: Server.handleClose: req: %+v\n", req)
	}

	err = Send(conn, packet)
	if err != nil {
		log.Println("websocket: server.handleClose: Send: " + err.Error())
	}

	serv.ClientRemove(conn)
}

// handleBadRequest by sending Close frame with status.
func (serv *Server) handleBadRequest(conn int) {
	var (
		frameClose []byte = NewFrameClose(false, StatusBadRequest, nil)
		err        error
	)

	err = Send(conn, frameClose)
	if err != nil {
		log.Println("websocket: server.handleBadRequest: " + err.Error())
	}

	_, err = Recv(conn)
	if err != nil {
		log.Println("websocket: server.handleBadRequest: " + err.Error())
	}

	serv.ClientRemove(conn)
}

// handleInvalidData by sending Close frame with status 1007.
func (serv *Server) handleInvalidData(conn int) {
	var (
		frameClose []byte = NewFrameClose(false, StatusInvalidData, nil)
		err        error
	)

	err = Send(conn, frameClose)
	if err != nil {
		log.Println("websocket: server.handleInvalidData: " + err.Error())
	}

	_, err = Recv(conn)
	if err != nil {
		log.Println("websocket: server.handleInvalidData: " + err.Error())
	}

	serv.ClientRemove(conn)
}

// handlePing from client by sending pong response.
//
// “`RFC6455
// (5.5.3.P3)
// A Pong frame sent in response to a Ping frame must have identical
// "Application data" as found in the message body of the Ping frame
// being replied to.
// “`
func (serv *Server) handlePing(conn int, req *Frame) {
	if debug.Value >= 3 {
		fmt.Printf("websocket: Server.handlePing: conn:%d frame:%+v\n",
			conn, req)
	}

	req.opcode = OpcodePong
	req.masked = 0

	var (
		res []byte = req.pack()
		err error
	)

	err = Send(conn, res)
	if err != nil {
		log.Println("websocket: server.handlePing: " + err.Error())
		serv.ClientRemove(conn)
		return
	}
}

// reader read request from client.
//
// To avoid confusing network intermediaries (such as intercepting proxies)
// and for security reasons that are further discussed in Section 10.3, a
// client MUST mask all frames that it sends to the server (see Section 5.3
// for further details).  (Note that masking is done whether or not the
// WebSocket Protocol is running over TLS.)  The server MUST close the
// connection upon receiving a frame that is not masked.  In this case, a
// server MAY send a Close frame with a status code of 1002 (protocol error)
// as defined in Section 7.4.1. (RFC 6455, section 5.1, P27).
func (serv *Server) reader() {
	var (
		frames    *Frames
		frame     *Frame
		err       error
		packet    []byte
		fds       []int
		conn      int
		max       int
		x         int
		isClosing bool
	)

	for {
		fds, err = serv.poll.WaitRead()
		if err != nil {
			log.Println("websocket: Server.reader: " + err.Error())
			break
		}

		for x = 0; x < len(fds); x++ {
			conn = fds[x]

			packet, err = Recv(conn)
			if err != nil || len(packet) == 0 {
				serv.ClientRemove(conn)
				continue
			}

			if debug.Value >= 3 {
				max = len(packet)
				if max > 16 {
					max = 16
				}
				fmt.Printf("websocket: Server.reader: packet {len:%d value:% x ...}\n",
					len(packet), packet[:max])
			}

			// Handle chopped, unfinished packet or payload.
			frame, _ = serv.Clients.getFrame(conn)
			if frame != nil {
				packet = frame.unpack(packet)
				if frame.isComplete {
					serv.Clients.setFrame(conn, nil)
					isClosing = serv.handleFrame(conn, frame)
					if isClosing {
						continue
					}
				}
				if len(packet) == 0 {
					serv.poll.ReregisterRead(x, conn)
					continue
				}
			}

			frames = Unpack(packet)
			if frames == nil {
				serv.ClientRemove(conn)
				continue
			}

			if debug.Value >= 3 {
				fmt.Printf("websocket: Server.reader: frames: len:%d\n", len(frames.v))
			}

			var isClosing bool
			for _, frame = range frames.v {
				if !frame.isComplete {
					serv.Clients.setFrame(conn, frame)
					continue
				}

				isClosing = serv.handleFrame(conn, frame)
				if isClosing {
					break
				}
			}
			if !isClosing {
				serv.poll.ReregisterRead(x, conn)
			}
		}
	}
}

// pinger is a routine that send control PING frame to all client connections
// every N seconds.
func (serv *Server) pinger() {
	var (
		pingTicker *time.Ticker = time.NewTicker(16 * time.Second)
		framePing  []byte       = NewFramePing(false, nil)

		all  []int
		conn int
		err  error
	)

	for {
		select {
		case <-pingTicker.C:
			all = serv.Clients.All()

			for _, conn = range all {
				err = Send(conn, framePing)
				if err != nil {
					// Error on sending PING will be
					// assumed as bad connection.
					serv.ClientRemove(conn)
				}
			}
		case <-serv.running:
			return
		}
	}
}

// Start accepting incoming connection from clients.
func (serv *Server) Start() (err error) {
	err = serv.createSockServer()
	if err != nil {
		return
	}

	serv.chUpgrade = make(chan int, _maxQueue)
	go serv.upgrader()

	serv.poll, err = libnet.NewPoll()
	if err != nil {
		return
	}

	go serv.reader()
	go serv.pinger()

	var conn int
	for {
		conn, _, err = unix.Accept(serv.sock)
		if err != nil {
			if err.Error() == "software caused connection abort" {
				// Stop has been called
				return nil
			}
			log.Println("websocket: unix.Accept: " + err.Error())
			return
		}

		serv.chUpgrade <- conn
	}
}

// Stop the server.
func (serv *Server) Stop() {
	var err error = unix.Close(serv.sock)
	if err != nil {
		log.Println("websocket: Stop: unix.Close: " + err.Error())
	}

	serv.running <- struct{}{}

	serv.poll.Close()

	close(serv.chUpgrade)
}

// sendResponse to client.
func (serv *Server) sendResponse(conn int, res *Response) (err error) {
	var (
		packet []byte
	)

	packet, err = json.Marshal(res)
	if err != nil {
		log.Println("websocket: server.sendResponse: " + err.Error())
		return
	}

	packet = NewFrameText(false, packet)

	err = Send(conn, packet)
	if err != nil {
		log.Println("websocket: server.sendResponse: " + err.Error())
	}

	return
}
