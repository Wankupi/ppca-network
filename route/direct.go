package route

import (
	"fmt"
	"main/outbound"
	"net"
)

type DirectRouter struct {
}

func (dr DirectRouter) RoutingDomain(addr string, port uint16) (outbound.OutboundConnTCP, error) {
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", addr, port))
	if err != nil {
		return nil, err
	}
	return dr.RoutingIP(laddr.IP, port)
}

func (dr DirectRouter) RoutingIP(ip net.IP, port uint16) (outbound.OutboundConnTCP, error) {
	return outbound.NewDirectTCPwithIP(ip, port)
}

func (dr DirectRouter) RoutingUDP() (outbound.OutboundConnUDP, error) {
	return outbound.NewDirectUDP()
}

func (dr DirectRouter) RoutingRawIP() (outbound.OutboundConnRawIP, error) {
	return outbound.NewDirectIP()
}
