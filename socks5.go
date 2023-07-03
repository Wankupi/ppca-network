package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

func main() {
	listen, err := net.Listen("tcp", "0.0.0.0:10000")
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
		go process(conn)
	}
}

func process(conn net.Conn) {
	reader := bufio.NewReader(conn)
	var buf [260]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		fmt.Printf("read from connect failed, error code = %v\n", err)
		return
	}
	recv := string(buf[:n])

	if recv[0] != 0x05 || recv[1] == 0 {
		fmt.Print("version is wrong or nMethods is 0\n")
		conn.Close()
		return
	}

	conn.Write([]byte{0x05, 0x00}) // 不需要认证

	n, err = reader.Read(buf[:])

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

	fmt.Printf("%v: %v\n", time.Now().Format(time.RFC3339), addr)
	defer fmt.Print("close\n")

	d := net.Dialer{Timeout: time.Duration(5) * time.Second}
	server, err := d.Dial("tcp", addr)
	if err != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		conn.Close()
		return
	}

	localIP_str, localPort_str, err := net.SplitHostPort(server.LocalAddr().String())
	localIP := net.ParseIP(localIP_str)
	localPort, _ := strconv.Atoi(localPort_str)

	var ATYP byte = 0x01
	if len(localIP) == 16 {
		ATYP = 0x04
	}
	res := []byte{0x05, 0x00, 0x00, ATYP}
	res = append(res, localIP...)
	conn.Write(binary.BigEndian.AppendUint16(res, uint16(localPort)))
	go func() {
		defer server.Close()
		defer conn.Close()
		io.CopyBuffer(server, conn, make([]byte, 1024))
	}()
	io.CopyBuffer(conn, server, make([]byte, 1024))
}
