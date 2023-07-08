package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type socksTCP struct {
	client *net.TCPConn
	server *net.TCPConn
}

func newSocksConnTCP(client *net.TCPConn, addr string, port uint16) (socksConn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr+fmt.Sprintf(":%v", port))
	var server *net.TCPConn
	if err == nil {
		server, err = net.DialTCP("tcp", nil, raddr)
	}
	if err != nil {
		var fail_type byte
		if strings.Contains(err.Error(), "refused") {
			fail_type = 0x05
		} else if strings.Contains(err.Error(), "no route") {
			fail_type = 0x03
		} else if strings.Contains(err.Error(), "i/o timeout") {
			fail_type = 0x04
		} else if strings.Contains(err.Error(), "network is unreachable") {
			fail_type = 0x03
		} else if strings.Contains(err.Error(), "failure in name resolution") {
			fail_type = 0x04
		}
		client.Write([]byte{0x05, fail_type})
		return nil, errors.New("dial failed, code: " + err.Error())
	}
	sendBackAddr(client, server.LocalAddr())
	if err != nil {
		return nil, errors.New("send back addr failed, code: " + err.Error())
	}

	return &socksTCP{client: client, server: server}, nil
}

func (socks *socksTCP) run(ctx context.Context) {
	A := socks.client
	B := socks.server
	go func() {
		io.CopyBuffer(A, B, make([]byte, 1024))
		A.CloseWrite()
		B.CloseRead()
	}()
	io.CopyBuffer(B, A, make([]byte, 1024))
	B.CloseWrite()
	A.CloseRead()
}
