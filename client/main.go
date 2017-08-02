package main

import (
	"runtime"
	"net"
	"fmt"
	"tcp_tunnel/config"
	"flag"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var ip string
	var port string
	flag.StringVar(&ip, "i", "127.0.0.1", "local ip")
	flag.StringVar(&port, "p", "8888", "local port")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		panic("listen ip is nil")
	}

}