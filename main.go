package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
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
	defer listen.Close()
	fmt.Printf("listening on port %v\n", port)
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Printf("accept failed, error code = %v\n", err)
			continue
		}
		go StartConnection(conn)
	}
}
