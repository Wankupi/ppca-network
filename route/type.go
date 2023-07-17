package route

import (
	"main/outbound"
	"net"
)

type Route interface {
	RoutingDomain(addr string, port uint16) (outbound.OutboundConnTCP, error)
	RoutingIP(ip net.IP, port uint16) (outbound.OutboundConnTCP, error)
	RoutingUDP() (outbound.OutboundConnUDP, error)
	RoutingRawIP() (outbound.OutboundConnRawIP, error)
}
