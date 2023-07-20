package tun

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"main/inbound"
	"main/route"
	"os"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/unix"
)

type tunServer struct {
	tun      *os.File
	router   route.Route
	tcpConns map[[16 + 16 + 2 + 2]byte](*tcpConn)
}

func Listen(deviceName string, router route.Route) (inbound.InboundServer, error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, errors.New("error open tun device, code: " + err.Error())
	}
	ifr, err := unix.NewIfreq(deviceName)
	if err != nil {
		return nil, err
	}
	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, file.Fd(), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(ifr)))
	if errno != 0 {
		return nil, errors.New("error ioctl(TunSetIff), code: " + unix.ErrnoName(errno))
	}
	return &tunServer{file, router, make(map[[36]byte]*tcpConn)}, nil
}

func (ts *tunServer) Accept(ctx context.Context) {
	buff := make([]byte, 65536)
	tun := ts.tun

	go func() {
		<-ctx.Done()
		tun.Close()
		fmt.Print("tun has been closed.\n")
	}()
	for {
		n, err := tun.Read(buff)
		if err != nil {
			fmt.Println("error read tun, code : ", err.Error())
			break
		}
		ts.dealIPv4(buff[:n])
	}
	tun.Close()
	for _, val := range ts.tcpConns {
		val.Close()
	}
	ts.tcpConns = nil
}

func (ts *tunServer) dealIPv4(data []byte) {
	packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

	ipv4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		fmt.Println("not tcp")
		return
	}

	var key [36]byte
	copy(key[:32], append(ipv4.SrcIP, ipv4.DstIP...))
	copy(key[32:34], binary.BigEndian.AppendUint16([]byte{}, uint16(tcp.SrcPort)))
	copy(key[32:34], binary.BigEndian.AppendUint16([]byte{}, uint16(tcp.DstPort)))

	conn, exist := ts.tcpConns[key]
	if !exist {
		conn = &tcpConn{SrcIP: ipv4.SrcIP, DstIP: ipv4.DstIP,
			SrcPort: uint16(tcp.SrcPort), DstPort: uint16(tcp.DstPort)}
	}

	tcp_back := ts.dealTCP(conn, tcp, key)
	if tcp_back == nil {
		return
	}
	send_tcp(ipv4.DstIP, ipv4.SrcIP, tcp_back)
}
