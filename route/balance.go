package route

import (
	"errors"
	"main/config"
	"main/outbound"
	socks5_out "main/outbound/socks5"
)

type BalanceRouter struct {
	outs []config.OutboundConfig
	k    int
}

func NewBalanceRouter(outs []config.OutboundConfig) *BalanceRouter {
	return &BalanceRouter{outs, 0}
}

func (router *BalanceRouter) RoutingTCP(addr string, port uint16) (outbound.OutboundConnTCP, error) {
	router.k = (router.k + 1) % len(router.outs)
	if router.outs[router.k].Protocol == "socks5" {
		return socks5_out.NewSocks5TCP(true, addr, port, router.outs[router.k].Addrs)
	} else if router.outs[router.k].Protocol == "direct" {
		return outbound.NewDirectTCP(addr, port)
	}
	return nil, errors.New("outbound protocol not support.")
}

func (router *BalanceRouter) RoutingUDP(addr string, port uint16) (outbound.OutboundConnUDP, error) {
	return outbound.NewDirectUDP()
}

func (router *BalanceRouter) RoutingRawIP(addr string, port uint16) (outbound.OutboundConnRawIP, error) {
	return outbound.NewDirectIP()
}
