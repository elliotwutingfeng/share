// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"net"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// TCPClient for DNS with TCP connection and list of remote addresses.
//
type TCPClient struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	addr         *net.TCPAddr
	conn         net.Conn
}

//
// NewTCPClient will create new DNS client with TCP network connection.
//
// The nameserver contains the IP address, not host name, of parent DNS
// server.  Default port is 53, if not set.
//
func NewTCPClient(nameserver string) (*TCPClient, error) {
	_, remoteIP, remotePort := libnet.ParseIPPort(nameserver, DefaultPort)
	if remoteIP == nil {
		return nil, fmt.Errorf("dns: invalid address '%s'", nameserver)
	}

	raddr := &net.TCPAddr{
		IP:   remoteIP,
		Port: int(remotePort),
	}

	cl := &TCPClient{
		readTimeout:  clientTimeout,
		writeTimeout: clientTimeout,
		addr:         raddr,
	}

	err := cl.Connect(raddr)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

//
// Close client connection.
//
func (cl *TCPClient) Close() error {
	if cl.conn != nil {
		return cl.conn.Close()
	}
	return nil
}

//
// Connect to remote address.
//
func (cl *TCPClient) Connect(raddr *net.TCPAddr) (err error) {
	laddr := &net.TCPAddr{IP: nil, Port: 0}

	cl.conn, err = net.DialTCP("tcp", laddr, raddr)

	return
}

//
// Lookup will query one of the name server with specific type, class, and
// name in synchronous mode.
//
// Name could be a host name for standard query or IP address for inverse
// query.
//
// This function is safe to be used concurrently.
//
func (cl *TCPClient) Lookup(
	allowRecursion bool, qtype, qclass uint16, qname string,
) (
	*Message, error,
) {
	if cl.addr == nil || cl.conn == nil {
		return nil, nil
	}
	if qtype == 0 {
		qtype = QueryTypeA
	}
	if qclass == 0 {
		qclass = QueryClassIN
	}

	msg := NewMessage()

	msg.Header.ID = getNextID()
	msg.Header.IsRD = allowRecursion
	msg.Header.QDCount = 1
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = qname

	_, _ = msg.Pack()

	res, err := cl.Query(msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// Query send DNS query to name server.
// The addr parameter is unused.
//
func (cl *TCPClient) Query(msg *Message) (*Message, error) {
	_, err := cl.Write(msg.Packet)
	if err != nil {
		return nil, err
	}

	res := NewMessage()

	_, err = cl.recv(res)
	if err != nil {
		return nil, err
	}

	err = res.Unpack()
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *TCPClient) RemoteAddr() string {
	return cl.addr.String()
}

//
// SetRemoteAddr set the remote address for sending the packet.
//
func (cl *TCPClient) SetRemoteAddr(addr string) (err error) {
	cl.addr, err = net.ResolveTCPAddr("udp", addr)
	return
}

//
// SetTimeout for sending and receiving packet.
//
func (cl *TCPClient) SetTimeout(t time.Duration) {
	cl.readTimeout = t
	cl.writeTimeout = t
}

//
// Write raw DNS response message on active connection.
// This method is only used by server to write the response of query to
// client.
//
func (cl *TCPClient) Write(msg []byte) (n int, err error) {
	if cl.writeTimeout > 0 {
		err = cl.conn.SetWriteDeadline(time.Now().Add(cl.writeTimeout))
		if err != nil {
			return
		}
	}

	lenmsg := len(msg)
	packet := make([]byte, 0, 2+lenmsg)

	libbytes.AppendUint16(&packet, uint16(lenmsg))
	packet = append(packet, msg...)

	n, err = cl.conn.Write(packet)

	return
}

//
// recv will read DNS message from active connection in client into `msg`.
//
func (cl *TCPClient) recv(msg *Message) (n int, err error) {
	if cl.readTimeout > 0 {
		err = cl.conn.SetReadDeadline(time.Now().Add(cl.readTimeout))
		if err != nil {
			return
		}
	}

	packet := make([]byte, maxUDPPacketSize)

	n, err = cl.conn.Read(packet)
	if err != nil {
		return
	}
	if n == 0 {
		return
	}

	msg.Packet = libbytes.Copy(packet[2:n])

	if debug.Value >= 3 {
		libbytes.PrintHex(">>> TCPClient: recv: ", msg.Packet, 8)
	}

	return
}
