package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
)

func main() {
	root_ctx, cancelAll := context.WithCancel(context.Background())
	stop_chan := make(chan os.Signal)
	signal.Notify(stop_chan, os.Interrupt)

	port := "10000" // default
	if len(os.Args) >= 2 {
		port = os.Args[1]
	} else {
		fmt.Print("Usage: " + os.Args[0] + " <port>\nOn default port " + port + "\n")
	}
	listenAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:"+port)
	listen, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		fmt.Printf("listen failed, error code = %v\n", err)
		return
	}

	go func() {
		<-stop_chan
		cancelAll()
		listen.Close()
	}()

	fmt.Printf("listening on port %v\n", port)
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			fmt.Printf("accept failed, error code = %v\n", err)
			continue
		}
		go StartConnection(root_ctx, conn)
	}
}
