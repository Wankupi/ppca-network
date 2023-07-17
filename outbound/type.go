package outbound

import "net"

type OutboundConnTCP interface {
	LocalAddr() net.Addr
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	CloseRead() error
	CloseWrite() error
}

type OutboundConnUDP interface {
	LocalAddr() net.Addr
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	Close() error
}

type OutboundConnRawIP interface {
	Write(b []byte) error
}
