package route

import (
	"errors"
	"main/outbound"
	socks5_out "main/outbound/socks5"
	"net"
)

type Socks5Router struct {
	IP   net.IP
	Port uint16
}

func (router *Socks5Router) RoutingDomain(addr string, port uint16) (outbound.OutboundConnTCP, error) {
	return socks5_out.NewSocks5TCP(true, []byte(addr), port, router.IP, router.Port)
}

func (router *Socks5Router) RoutingIP(ip net.IP, port uint16) (outbound.OutboundConnTCP, error) {
	return socks5_out.NewSocks5TCP(false, ip, port, router.IP, router.Port)
}

func (router *Socks5Router) RoutingUDP() (outbound.OutboundConnUDP, error) {
	return nil, errors.New("TODO")
}

func (router *Socks5Router) RoutingRawIP() (outbound.OutboundConnRawIP, error) {
	return outbound.NewDirectIP()
}
