package route

import (
	"main/outbound"
)

type Route interface {
	RoutingTCP(addr string, port uint16) (outbound.OutboundConnTCP, error)
	RoutingUDP(addr string, port uint16) (outbound.OutboundConnUDP, error)
	RoutingRawIP(addr string, port uint16) (outbound.OutboundConnRawIP, error)
}
