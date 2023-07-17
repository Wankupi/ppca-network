package outbound

import (
	"net"
	"syscall"
)

type DirectTCP = *net.TCPConn

func NewDirectTCPwithIP(ip net.IP, port uint16) (DirectTCP, error) {
	addr := net.TCPAddr{IP: ip, Port: int(port)}
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type DirectUDP = *net.UDPConn

func NewDirectUDP() (DirectUDP, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type DirectRawIP struct {
	fd int
}

var directIP *DirectRawIP

func NewDirectIP() (*DirectRawIP, error) {
	if directIP == nil {
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
		if err != nil {
			return nil, err
		}
		directIP = &DirectRawIP{fd: fd}
		err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_MARK, 0x111)
		if err != nil {
			syscall.Close(fd)
			return nil, err
		}
	}
	return directIP, nil
}

func (d *DirectRawIP) Write(b []byte) error {
	var addr syscall.SockaddrInet4
	addr.Addr = [4]byte{0, 0, 0, 0}
	addr.Port = 0
	err := syscall.Sendto(d.fd, b, 0, &addr)
	if err != nil {
		return err
	}
	return nil
}
