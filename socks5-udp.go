package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type socksUDP struct {
	client     *net.TCPConn
	server     *net.UDPConn
	relay      *net.UDPConn
	isLimitIP  bool
	ip         net.IP
	port       uint16
	clientAddr *net.UDPAddr
	relayAddr  *net.UDPAddr
}

func newSocksConnUDP(client *net.TCPConn, addr string, port uint16) (socksConn, error) {
	var udp_addr = &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0, Zone: ""}
	server, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		client.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("udp listen failed. code: " + err.Error())
	}

	// fmt.Print("udp listen from ", addr, " ", port, " -- local=", server.LocalAddr(), "\n")

	err = sendBackAddr(client, server.LocalAddr())
	if err != nil {
		server.Close()
		return nil, errors.New("send back addr failed, code: " + err.Error())
	}
	var udp socksUDP
	udp.client = client
	udp.server = server
	ip := net.ParseIP(addr)
	if ip != nil {
		udp.isLimitIP = true
		udp.ip = ip
	} else {
		udp.isLimitIP = false
	}
	udp.port = port
	return &udp, nil
}

func (socks *socksUDP) run() {
	defer socks.client.Close()
	defer socks.server.Close()
	var server = socks.server
	var buf [1024]byte
	for {
		n, laddr, err := server.ReadFromUDP(buf[:])
		if err != nil {
			fmt.Printf("read from udp client failed, code: %v\n", err.Error())
			continue
		}
		if socks.isLimitIP {
			if (!bytes.Equal(socks.ip, laddr.IP.To16())) || (socks.port != 0 && socks.port != laddr.AddrPort().Port()) {
				fmt.Print("udp client not wanted.")
				continue
			}
		}
		socks.clientAddr = laddr
		socks.send_udp(buf[:n], laddr)
	}
}

func (socks *socksUDP) send_udp(msg []byte, laddr net.Addr) {
	if msg[0] != 0x00 || msg[1] != 0x00 {
		fmt.Print("udp: RSV wrong, not 0x0000.\n")
		return
	}
	if msg[2] != 0x00 {
		fmt.Print("udp: FRAG not supported.\n")
		return
	}

	var to_addr string
	var index int
	switch msg[3] {
	case 0x01:
		to_addr = fmt.Sprintf("%v.%v.%v.%v", msg[4], msg[5], msg[6], msg[7])
		index = 8
		break
	case 0x03:
		to_addr = string(msg[5 : int(msg[4])+5])
		index = int(msg[4]) + 5
		break
	case 0x04:
		to_addr = "["
		for i := 4; i < 20; i += 2 {
			if i != 4 {
				to_addr += ":"
			}
			to_addr += fmt.Sprintf("%x", binary.BigEndian.Uint16(msg[i:i+2]))
		}
		to_addr += "]"
		index = 20
		break
	default:
		fmt.Print("udp: ATYP not supported.\n")
		return
	}
	to_port := binary.BigEndian.Uint16(msg[index : index+2])
	index += 2
	to_addr += fmt.Sprintf(":%v", to_port)

	rAddr, err := net.ResolveUDPAddr("udp", to_addr)

	if err != nil {
		fmt.Printf("udp: resolve remote error, code: %v\n", err.Error())
		return
	}

	if socks.relay == nil {
		socks.relay, err = net.ListenUDP("udp", nil)
		if err != nil {
			fmt.Printf("udp: dial remote error, code: %v\n", err.Error())
			return
		}
	}

	host := socks.relay

	m, err := host.WriteToUDP(msg[index:], rAddr)
	if err != nil {
		fmt.Printf("udp: host write fail, \033[31m m=%v  code:%v\033[0m\n", m, err.Error())
	}
	if host != nil && socks.relayAddr == nil {
		// on first send
		socks.relayAddr = new(net.UDPAddr)
		HostStr, PortStr, _ := net.SplitHostPort(host.LocalAddr().String())
		socks.relayAddr.IP = net.ParseIP(HostStr)
		socks.relayAddr.Port, _ = strconv.Atoi(PortStr)
		go socks.recv_udp()
	}
}

func (socks *socksUDP) recv_udp() {
	buf := make([]byte, 4096)
	for {
		n, raddr, err := socks.relay.ReadFromUDP(buf)
		if err != nil {
			break
		}
		msg := []byte{0x00, 0x00, 0x00}
		if raddr.IP.To4() != nil {
			msg = append(msg, 0x01)
			msg = append(msg, raddr.IP.To4()...)
		} else {
			msg = append(msg, 0x03)
			msg = append(msg, raddr.IP.To16()...)
		}
		msg = binary.BigEndian.AppendUint16(msg, uint16(raddr.Port))
		msg = append(msg, buf[:n]...)
		socks.server.WriteToUDP(msg, socks.clientAddr)
	}
}
