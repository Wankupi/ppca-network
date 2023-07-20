package main

import (
	"context"
	"fmt"
	"main/config"
	"main/inbound"
	"main/route"
	"os"
	"os/signal"
)

var Inbounds map[string]inbound.InboundServer
var Routers map[string]route.Route
var Outbounds map[string]config.OutboundConfig

func main() {
	con, err := config.GetConfig("config.json")
	if err != nil {
		fmt.Println("load config error. ", err.Error())
		return
	}

	Inbounds = make(map[string]inbound.InboundServer)
	Routers = make(map[string]route.Route)
	Outbounds = make(map[string]config.OutboundConfig)

	for _, oc := range con.Outbound {
		Outbounds[oc.Tag] = oc
	}
	for _, rc := range con.Route {
		var ocs []config.OutboundConfig
		for _, s := range rc.Outs {
			ocs = append(ocs, Outbounds[s])
		}
		Routers[rc.Tag] = route.NewBalanceRouter(ocs)
	}

	root_ctx, cancelAll := context.WithCancel(context.Background())
	stop_chan := make(chan os.Signal)
	signal.Notify(stop_chan, os.Interrupt)

	for _, ic := range con.Inbound {
		rou, exi := Routers[ic.Route]
		if !exi {
			fmt.Println("router [", ic.Route, "] not found, on [", ic.Tag, "]")
		}
		server, err := NewInboundServer(ic, rou)
		if err != nil {
			fmt.Printf("failed to start server [ %v ]\n", ic.Tag)
			continue
		}
		go server.Accept(root_ctx)
	}

	<-stop_chan
	cancelAll()
}
