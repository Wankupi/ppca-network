package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type socksUDP struct {
	client     *net.TCPConn
	local      *net.UDPConn
	remote     *net.UDPConn
	isLimitIP  bool
	ip         net.IP
	port       uint16
	clientAddr *net.UDPAddr
	relayAddr  *net.UDPAddr
}

func newSocksConnUDP(client *net.TCPConn, addr string, port uint16) (socksConn, error) {
	local, err := net.ListenUDP("udp", nil)
	if err != nil {
		client.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("udp: error on setup udp to local, code: " + err.Error())
	}

	remote, err := net.ListenUDP("udp", nil)
	if err != nil {
		local.Close()
		client.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil, errors.New("udp: error on setup udp to remote, code: " + err.Error())
	}

	sendBackAddr(client, local.LocalAddr())

	var socks socksUDP
	socks.client = client
	socks.local = local
	socks.remote = remote
	ip := net.ParseIP(addr)
	if ip != nil {
		socks.isLimitIP = true
		socks.ip = ip
	} else {
		socks.isLimitIP = false
	}
	socks.port = port

	socks.relayAddr = new(net.UDPAddr)
	HostStr, PortStr, _ := net.SplitHostPort(remote.LocalAddr().String())
	socks.relayAddr.IP = net.ParseIP(HostStr)
	socks.relayAddr.Port, _ = strconv.Atoi(PortStr)

	return &socks, nil
}

type udp_msg struct {
	msg   []byte
	rAddr *net.UDPAddr
}

func (socks *socksUDP) run(ctx context.Context) {
	defer socks.client.Close()
	defer socks.local.Close()
	defer socks.remote.Close()
	my_ctx, cancel := context.WithCancel(ctx)
	msg_to_remote := make(chan udp_msg, 1)
	msg_to_local := make(chan udp_msg, 1)
	go socks.recieveFromLocal(msg_to_remote, make([]byte, 1024))
	go socks.recieveFromRemote(msg_to_local, make([]byte, 1024))
	go func() {
		buf := make([]byte, 32)
		for {
			_, err := socks.client.Read(buf)
			if err != nil {
				break
			}
		}
		cancel()
	}()
	for {
		var err error
		done := false
		select {
		case msg := <-msg_to_local:
			_, err = socks.local.WriteToUDP(msg.msg, msg.rAddr)
		case msg := <-msg_to_remote:
			_, err = socks.remote.WriteToUDP(msg.msg, msg.rAddr)
		case <-my_ctx.Done():
			done = true
		}
		if err != nil {
			fmt.Println(err.Error())
		}
		if done {
			break
		}
	}
}

func (socks *socksUDP) recieveFromLocal(remote_chan chan udp_msg, buf []byte) {
	for {
		n, laddr, err := socks.local.ReadFromUDP(buf)
		if err != nil {
			break
		}
		if socks.isLimitIP {
			if (!bytes.Equal(socks.ip, laddr.IP.To16())) || (socks.port != 0 && socks.port != laddr.AddrPort().Port()) {
				fmt.Print("udp client not wanted.")
				continue
			}
		}
		socks.clientAddr = laddr // assume there is only one client
		socks.dealRecvLocalMsg(buf[:n], laddr, remote_chan)
	}
	close(remote_chan)
}

func (socks *socksUDP) dealRecvLocalMsg(msg []byte, laddr net.Addr, remote_chan chan udp_msg) {
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
	remote_chan <- udp_msg{msg[index:], rAddr}
}

func (socks *socksUDP) recieveFromRemote(local_chan chan udp_msg, buf []byte) {
	for {
		n, raddr, err := socks.remote.ReadFromUDP(buf)
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
		local_chan <- udp_msg{msg, socks.clientAddr}
	}
	close(local_chan)
}
