package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
	"tcp_tunnel/logic"
)

var (
	localIp    = flag.String("i", "127.0.0.1", "local ip addr")
	localPort  = flag.String("p", "8888", "local port")
	remoteIp   = flag.String("ri", "127.0.0.1", "remote ip addr")
	remotePort = flag.String("rp", "8787", "remote port")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", localIp, localPort))
	if err != nil {
		panic("listen ip is nil")
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic("accept error.")
		}
		go transmit(conn)
	}
}

func transmit(tcpConn *net.Conn) {
	for {
		data, err := ioutil.ReadAll(tcpConn)
		if err != nil {
			errors.New("read data error.")
		}
		tip := logic.NewTipBuffer()
		transmitBuff := tip.TransmitStream(remoteIp, remotePort, data)

		remoteTcpAddr, err := net.ResolveTCPAddr("tcp4", remoteIp+":"+remotePort)
		if err != nil {
			panic("resolve remote tcp addr failed.")
			return
		}
		remoteConn, err := net.DialTCP("tcp", nil, remoteTcpAddr)
		if err != nil {
			panic("dial remote tcp failed.")
			return
		}
		_, err = remoteConn.Write(transmitBuff)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}
