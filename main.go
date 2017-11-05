package main

import (
	"flag"
	"runtime"
	"tcp_tunnel/config"
	"tcp_tunnel/server"
)

var (
	ServerIp, ServerPort string
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&ServerIp, "i", "0.0.0.0", "server ip")
	flag.StringVar(&ServerPort, "p", "8787", "server port")
	flag.Parse()
}

func main() {
	conf = config.NewConfig(ServerIp, ServerPort)
	go server.TunnelListen(conf)
	select {}
}
