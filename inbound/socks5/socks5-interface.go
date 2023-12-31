package socks5

import (
	"context"
	"errors"
	"net"
)

type socksConn interface {
	run(ctx context.Context)
}

func (server *Socks5Listener) newSocksConn(conn_type string, client *net.TCPConn, addr string, port uint16) (socksConn, error) {
	if conn_type == "tcp" {
		return server.newSocksConnTCP(client, addr, port)
	} else if conn_type == "bind" {
		client.Write([]byte{0x05, 0x07})
		return nil, errors.New("unsupported bind command now")
	} else if conn_type == "udp" {
		return server.newSocksConnUDP(client, addr, port)
	} else {
		client.Write([]byte{0x05, 0x07})
		return nil, errors.New("unsupported command")
	}
}
