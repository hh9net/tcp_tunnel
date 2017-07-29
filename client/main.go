package main

import (
	"runtime"
	"net"
	"fmt"
	"tcp_tunnel/config"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ListenIp, config.ListenPort))
	if err != nil {
		panic("listen ip is nil")
	}
}