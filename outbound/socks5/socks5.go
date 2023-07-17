package socks5_out

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type SocksTCP = *net.TCPConn

func NewSocks5TCP(isDomain bool, dest net.IP, port uint16, proxy_ip net.IP, proxy_port uint16) (SocksTCP, error) {
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: proxy_ip, Port: int(proxy_port)})
	if err != nil {
		return nil, err
	}
	conn.Write([]byte{0x05, 0x01, 0x00})
	if isDomain {
		conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x03, byte(len(dest))}, dest...), port))
	} else if dest.To4() != nil {
		conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x01}, dest[len(dest)-4:]...), port))
	} else {
		conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x04}, dest...), port))
	}
	buf := make([]byte, 300)
	_, err = io.ReadFull(conn, buf[:2])
	if err != nil || buf[0] != 0x05 || buf[1] != 0x00 {
		return nil, errors.New("socks error.")
	}
	_, err = io.ReadFull(conn, buf[:5])
	if err != nil || buf[0] != 0x05 || buf[1] != 0x00 || buf[2] != 0x00 {
		return nil, errors.New("socks error.")
	}
	switch buf[3] {
	case 0x01:
		_, err = io.ReadFull(conn, buf[5:10])
	case 0x03:
		_, err = io.ReadFull(conn, buf[5:7+buf[4]])
	case 0x04:
		_, err = io.ReadFull(conn, buf[5:22])
	default:
		return nil, errors.New("socks error.")
	}
	if err != nil {
		return nil, errors.New("socks error.")
	}
	return conn, nil
}
