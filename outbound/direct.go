package outbound

import "net"

type DirectTCP = *net.TCPConn

func NewDirectTCPwithIP(ip net.IP, port uint16) (DirectTCP, error) {
	addr := net.TCPAddr{IP: ip, Port: int(port)}
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type DirectUDP = *net.UDPConn

func NewDirectUDPwithIP() (DirectUDP, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
