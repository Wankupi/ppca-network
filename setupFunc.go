package main

import (
	"errors"
	"main/config"
	"main/inbound"
	"main/inbound/socks5"
	"main/inbound/tun"
	"main/route"
)

func NewInboundServer(con config.InboundConfig, router route.Route) (inbound.InboundServer, error) {
	switch con.Protocol {
	case "socks5":
		return socks5.NewSock5Listner(con.Listen, router)
	case "tun":
		return tun.Listen(con.Listen, router)
	}
	return nil, errors.New("inbound protocol not supported")
}
