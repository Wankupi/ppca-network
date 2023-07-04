package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	port := "10000" // default
	if len(os.Args) >= 2 {
		port = os.Args[1]
	} else {
		fmt.Print("Usage: " + os.Args[0] + " <port>\nOn default port " + port + "\n")
	}
	listen, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Printf("listen failed, error code = %v\n", err)
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("accept failed, error code = %v\n", err)
			continue
		}
		go start_connection(conn)
	}
}

func start_connection(conn net.Conn) {
	defer conn.Close()
	server, err := hand_shake(conn)
	if err != nil {
		return
	}
	defer server.Close()
	go io.CopyBuffer(server, conn, make([]byte, 1024))
	io.CopyBuffer(conn, server, make([]byte, 1024))
}

func hand_shake(conn net.Conn) (sever net.Conn, err error) {
	var buf [260]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}
	if buf[0] != 0x05 || buf[1] == 0 {
		return nil, errors.New("version is wrong or nMethods is 0")
	}
	conn.Write([]byte{0x05, 0x00}) // 不需要认证

	n, err = conn.Read(buf[:])
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}

	port := binary.BigEndian.Uint16(buf[n-2 : n])
	addr := ""
	switch buf[3] {
	case 0x01: // ipv4
		addr = fmt.Sprintf("%v.%v.%v.%v:%v", buf[4], buf[5], buf[6], buf[7], port)
		break
	case 0x03: // domain
		addr = string(buf[5:n-2]) + fmt.Sprintf(":%v", port)
		break
	case 0x04: // ipv6
		addr = "["
		for i := 4; i < 20; i += 2 {
			if i != 4 {
				addr += ":"
			}
			addr += fmt.Sprintf("%x", binary.BigEndian.Uint16(buf[i:i+2]))
		}
		addr += fmt.Sprintf("]:%v", port)
		break
	}

	fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05.000"), addr)

	d := net.Dialer{Timeout: time.Duration(5) * time.Second}
	server, err := d.Dial("tcp", addr)
	if err != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("dial failed, code: " + err.Error())
	}

	localIP_str, localPort_str, err := net.SplitHostPort(server.LocalAddr().String())
	localIP := net.ParseIP(localIP_str)
	localPort, _ := strconv.Atoi(localPort_str)

	res := []byte{0x05, 0x00, 0x00, 0x04}
	res = append(res, localIP...)
	conn.Write(binary.BigEndian.AppendUint16(res, uint16(localPort)))
	return server, nil
}
