package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type socksTCP struct {
	client net.Conn
	server net.Conn
}

func newSocksConnTCP(client net.Conn, addr string, port uint16) (socksConn, error) {
	server, err := net.DialTimeout("tcp", addr+":"+fmt.Sprint(port), time.Duration(5)*time.Second)
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

	err = sendBackAddr(client, server.LocalAddr())
	if err != nil {
		return nil, errors.New("send back addr failed, code: " + err.Error())
	}

	return &socksTCP{client: client, server: server}, nil
}

func (socks *socksTCP) run() {
	forward_and_close(socks.client, socks.server)
}
