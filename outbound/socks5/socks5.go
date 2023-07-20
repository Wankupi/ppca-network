package socks5_out

import (
	"encoding/binary"
	"errors"
	"io"
	"main/config"
	"net"
)

type SocksTCP = *net.TCPConn

func NewSocks5TCP(IsDomain bool, Dest string, Port uint16, proxy []config.Addr) (SocksTCP, error) {
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(proxy[0].IP), Port: int(proxy[0].Port)})
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 300)
	for i := 1; i <= len(proxy); i += 1 {
		conn.Write([]byte{0x05, 0x01, 0x00})
		var dIP string
		var dport uint16
		if i < len(proxy) {
			dIP = proxy[i].IP
			dport = proxy[i].Port
		} else {
			dIP = Dest
			dport = Port
		}
		dest := net.ParseIP(dIP)
		if dest == nil {
			conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x03, byte(len(dIP))}, dIP...), dport))
		} else if dest.To4() != nil {
			conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x01}, dest[len(dest)-4:]...), dport))
		} else {
			conn.Write(binary.BigEndian.AppendUint16(append([]byte{0x05, 0x01, 0x00, 0x04}, dest...), dport))
		}
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
	}
	return conn, nil
}
