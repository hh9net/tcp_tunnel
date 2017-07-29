package main

import (
	"runtime"
	"tcp_tunnel/server"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	go server.TunnelListen()
	select {}
}
