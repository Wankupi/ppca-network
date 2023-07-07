package main

import (
	"errors"
	"net"
)

type socksConn interface {
	run()
}

func newSocksConn(conn_type string, client net.Conn, addr string, port uint16) (socksConn, error) {
	if conn_type == "tcp" {
		return newSocksConnTCP(client, addr, port)
	} else if conn_type == "bind" {
		client.Write([]byte{0x05, 0x07})
		return nil, errors.New("unsupported bind command now")
	} else if conn_type == "udp" {
		return newSocksConnUDP(client, addr, port)
	} else {
		client.Write([]byte{0x05, 0x07})
		return nil, errors.New("unsupported command")
	}
}