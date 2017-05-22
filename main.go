package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"tcp_tunnel/server"
)

func main() {
	go server.StartServer()
	select {}
}
