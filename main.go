package main

import (
	"context"
	"fmt"
	"main/socks5"
	"os"
	"os/signal"
	"strconv"
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
	Port, _ := strconv.Atoi(port)

	go func() {
		<-stop_chan
		cancelAll()
	}()

	s5, err := socks5.Listen("0.0.0.0", uint16(Port))
	if err != nil {
		fmt.Println("failed to listen socks5, code: ", err.Error())
		return
	}
	s5.Accept(root_ctx)
}
