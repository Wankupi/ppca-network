package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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
		fmt.Print("\033[31m", err, "\033[0m\n")
		return
	}
	defer server.Close()
	go io.CopyBuffer(server, conn, make([]byte, 1024))
	io.CopyBuffer(conn, server, make([]byte, 1024))
}

func hand_shake(conn net.Conn) (sever net.Conn, err error) {
	var buf [260]byte
	reader := bufio.NewReader(conn)
	// n, err := reader.Read(buf[:])
	n, err := io.ReadFull(reader, buf[:2])
	authSupport := false
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}
	// if buf[0] != 0x05 {
	// 	authSupport = true
	// }
	n, err = io.ReadFull(reader, buf[:buf[1]])
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

	var isUnsupport bool = false
	n, err = io.ReadFull(reader, buf[:5])
	if err != nil {
		return nil, errors.New("read from connect failed, code: " + err.Error())
	}
	if (buf[0] != 0x05) || (buf[1] != 0x01) {
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("version is wrong or command is not correct")
	}
	if buf[3] == 0x01 {
		m, err := io.ReadFull(reader, buf[5:10])
		if err != nil {
			return nil, errors.New("read from connect failed, code: " + err.Error())
		}
		n += m
	} else if buf[3] == 0x03 {
		m, err := io.ReadFull(reader, buf[5:7+buf[4]])
		if err != nil {
			return nil, errors.New("read from connect failed, code: " + err.Error())
		}
		n += m
	} else if buf[3] == 0x04 {
		m, err := io.ReadFull(reader, buf[5:22])
		if err != nil {
			return nil, errors.New("read from connect failed, code: " + err.Error())
		}
		n += m
	} else {
		isUnsupport = true
	}

	port := binary.BigEndian.Uint16(buf[n-2 : n])
	addr := ""
	switch buf[3] {
	case 0x01: // ipv4
		if n != 10 {
			isUnsupport = true
		}
		addr = fmt.Sprintf("%v.%v.%v.%v:%v", buf[4], buf[5], buf[6], buf[7], port)
		break
	case 0x03: // domain
		if n != int(buf[4])+7 {
			isUnsupport = true
		}
		addr = string(buf[5:n-2]) + fmt.Sprintf(":%v", port)
		break
	case 0x04: // ipv6
		if n != 22 {
			isUnsupport = true
		}
		addr = "["
		for i := 4; i < 20; i += 2 {
			if i != 4 {
				addr += ":"
			}
			addr += fmt.Sprintf("%x", binary.BigEndian.Uint16(buf[i:i+2]))
		}
		addr += fmt.Sprintf("]:%v", port)
		break
	default:
		isUnsupport = true
	}
	if isUnsupport {
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("unsupported address type")
	}

	fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05.000"), addr)

	d := net.Dialer{Timeout: time.Duration(5) * time.Second}
	server, err := d.Dial("tcp", addr)
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
		conn.Write([]byte{0x05, fail_type, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
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
