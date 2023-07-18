package tun

import (
	"fmt"
	"main/outbound"
	"math/rand"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func max(a uint32, b uint32) uint32 {
	if a < b {
		return b
	} else {
		return a
	}
}

func send_tcp(src net.IP, dst net.IP, pack *layers.TCP) {
	if dst.To4() != nil {
		send_tcp_ipv4(src, dst, pack)
	} else {
		fmt.Print("ipv6 not supported.")
	}
}

var ipv4_id uint16 = uint16(rand.Uint32())

func send_tcp_ipv4(src net.IP, dst net.IP, pack *layers.TCP) {
	ip_back := layers.IPv4{
		Version:    4,
		IHL:        0, // auto set
		TOS:        0,
		Length:     20 + uint16(len(pack.Contents)),
		Id:         ipv4_id,
		Flags:      layers.IPv4Flag(2),
		FragOffset: 0,
		TTL:        64,
		Protocol:   6,
		Checksum:   0, // auto calc
		SrcIP:      src,
		DstIP:      dst,
		Options:    nil,
		Padding:    nil}

	ip_back.Payload = pack.Contents
	pack.SetNetworkLayerForChecksum(&ip_back)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err := gopacket.SerializeLayers(buf, opts,
		&ip_back, pack, gopacket.Payload(pack.Payload))
	if err != nil {
		fmt.Println("\033[31m", "serialize error ", err.Error(), "\033[0m")
	}
	out, err := outbound.NewDirectIP()
	if err == nil {
		err = out.Write(buf.Bytes())
	}
	if err != nil {
		fmt.Println("\033[31m", "send back msg error: ", err.Error(), "\033[0m")
	}
}

type tcpConn struct {
	SrcIP                            net.IP
	SrcPort                          uint16
	DstIP                            net.IP
	DstPort                          uint16
	sendHead                         uint32
	recvHead                         uint32
	setup                            bool
	myFin, heFin, myAckFin, heAckFin bool
	relay                            outbound.OutboundConnTCP
}

func (t tcpConn) Close() {
	if t.relay != nil {
		t.relay.Close()
		t.relay = nil
	}
}

func (ts *tunServer) dealTCP(conn *tcpConn, pack *layers.TCP, key [36]byte) *layers.TCP {
	_, exist := ts.tcpConns[key]
	if !exist {
		if !pack.SYN {
			return nil
		}
		ts.tcpConns[key] = conn
	}
	need_reply := false
	conn.recvHead = pack.Seq + max(uint32(len(pack.Payload)), 1)
	backpack := layers.TCP{
		SrcPort:    pack.DstPort,
		DstPort:    pack.SrcPort,
		Seq:        conn.sendHead,
		Ack:        conn.recvHead,
		DataOffset: 0, // auto calc
		FIN:        false,
		SYN:        false,
		RST:        false,
		ACK:        true,
		URG:        false,
		ECE:        false,
		CWR:        false,
		NS:         false,
		Window:     64240,
		Checksum:   0, // auto calc
		Urgent:     0,
		Options:    nil,
		Padding:    nil}
	if pack.SYN {
		need_reply = true
		backpack.SYN = true
		backpack.Seq = rand.Uint32()
		conn.sendHead = backpack.Seq + 1
	}

	if pack.ACK {
		if !conn.setup {
			conn.setup = true
			var err error
			conn.relay, err = ts.router.RoutingIP(conn.DstIP, uint16(conn.DstPort))
			if err != nil {
				return nil
			}
			go ts.relayRemoteToLocal(conn)
		}
		if conn.myFin {
			conn.heAckFin = true
		}
	}

	if pack.FIN {
		backpack.FIN = true
		conn.heFin = true
		conn.myFin = true
		need_reply = !conn.myAckFin
		conn.myAckFin = true
	}

	if pack.Payload != nil && len(pack.Payload) > 0 {
		need_reply = true
		_, err := conn.relay.Write(pack.Payload)
		if err != nil {
			if !conn.myFin {
				backpack.FIN = true
				conn.myFin = true
			}
		}
	}

	if pack.RST {
		conn.heFin, conn.myFin, conn.heAckFin, conn.myAckFin = true, true, true, true
		need_reply = false
	}
	if conn.heFin && conn.myFin && conn.heAckFin && conn.myAckFin {
		conn.Close()
		delete(ts.tcpConns, key)
	}
	if need_reply {
		return &backpack
	} else {
		return nil
	}
}

func (ts *tunServer) relayRemoteToLocal(conn *tcpConn) {
	pack := layers.TCP{
		SrcPort:    layers.TCPPort(conn.DstPort),
		DstPort:    layers.TCPPort(conn.SrcPort),
		Seq:        conn.sendHead,
		Ack:        0,
		DataOffset: 0,
		FIN:        false,
		SYN:        false,
		RST:        false,
		ACK:        true,
		URG:        false,
		ECE:        false,
		CWR:        false,
		NS:         false,
		PSH:        false,
		Window:     64240,
		Checksum:   0, // auto calc
		Urgent:     0,
		Options:    nil,
		Padding:    nil}
	buf := make([]byte, 4096)
	for {
		n, err := conn.relay.Read(buf)
		if err != nil {
			break
		}
		pack.Payload = buf[:n]
		pack.Seq = conn.sendHead
		conn.sendHead += uint32(n)
		pack.Ack = conn.recvHead
		send_tcp(conn.DstIP, conn.SrcIP, &pack)
	}
	pack.FIN = true
	pack.Seq = conn.sendHead
	conn.sendHead += 1
	pack.Ack = conn.recvHead
	pack.Payload = nil
	if !conn.myFin {
		conn.myFin = true
		send_tcp(conn.DstIP, conn.SrcIP, &pack)
	}
}
