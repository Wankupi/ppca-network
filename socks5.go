package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

func StartConnection(ctx context.Context, conn *net.TCPConn) {
	socks, err := hand_shake(conn)
	if err != nil {
		fmt.Print("\033[31m", err, "\033[0m\n")
		conn.CloseWrite()
		conn.CloseRead()
		return
	}
	go socks.run(ctx)
}

func hand_shake(conn *net.TCPConn) (socks socksConn, err error) {
	var buf [260]byte

	n, err := io.ReadFull(conn, buf[:2])
	authSupport := false
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}

	n, err = io.ReadFull(conn, buf[:buf[1]])
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}
	for _, val := range buf[:n] {
		if val == 0x00 {
			authSupport = true
			break
		}
	}
	if !authSupport {
		conn.Write([]byte{0x05, 0xff}) // 不支持任何认证方式
		return nil, errors.New("no auth method is supported")
	}
	conn.Write([]byte{0x05, 0x00}) // 不需要认证

	n, err = io.ReadFull(conn, buf[:5])
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}

	if buf[0] != 0x05 { // check socks version
		conn.Write([]byte{0x05, 0x07})
		return nil, errors.New("version is wrong")
	}

	// fmt.Print("\033[33m", int(buf[1]), "\033[0m\n")
	var net_type string
	if buf[1] == 0x01 {
		net_type = "tcp"
	} else if buf[1] == 0x02 {
		net_type = "bind"
	} else if buf[1] == 0x03 {
		net_type = "udp"
	} else {
		conn.Write([]byte{0x05, 0x07})
		return nil, errors.New("unsupported command")
	}

	if buf[3] == 0x01 { // ipv4
		m, errr := io.ReadFull(conn, buf[5:10])
		err = errr
		n += m
	} else if buf[3] == 0x03 { // domain
		m, errr := io.ReadFull(conn, buf[5:7+buf[4]])
		err = errr
		n += m
	} else if buf[3] == 0x04 { // ipv6
		m, errr := io.ReadFull(conn, buf[5:22])
		err = errr
		n += m
	} else {
		conn.Write([]byte{0x05, 0x08})
		return nil, errors.New("unsupported address type")
	}
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}

	port := binary.BigEndian.Uint16(buf[n-2 : n])
	addr := ""
	switch buf[3] {
	case 0x01: // ipv4
		addr = fmt.Sprintf("%v.%v.%v.%v", buf[4], buf[5], buf[6], buf[7])
		break
	case 0x03: // domain
		addr = string(buf[5 : n-2])
		break
	case 0x04: // ipv6
		addr = "["
		for i := 4; i < 20; i += 2 {
			if i != 4 {
				addr += ":"
			}
			addr += fmt.Sprintf("%x", binary.BigEndian.Uint16(buf[i:i+2]))
		}
		addr += "]"
		break
	}
	fmt.Printf("[%v] %v:%v\n", time.Now().Format("15:04:05.000"), addr, port)

	return newSocksConn(net_type, conn, addr, port)
}

func sendBackAddr(conn *net.TCPConn, addr net.Addr) {
	localIP_str, localPort_str, _ := net.SplitHostPort(addr.String())
	localIP := net.ParseIP(localIP_str)
	localPort, _ := strconv.Atoi(localPort_str)
	res := []byte{0x05, 0x00, 0x00, 0x04}
	res = append(res, localIP...)
	conn.Write(binary.BigEndian.AppendUint16(res, uint16(localPort)))
}
