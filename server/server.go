package server

import (
	"fmt"
	"io"
	"net"
	"tcp_tunnel/config"
	"tcp_tunnel/logic"
)

func TunnelListen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ListenIp, config.ListenPort))
	if err != nil {
		panic("listen ip is nil")
	}
	for {
		tcpConntection, err := logic.Accept(listener)
		if err != nil {
			continue
		}
		go server(NewTunnel(tcpConntection))
	}
}

type Tunnel struct {
	tcpConnection *logic.TcpConnection
	tipBuffer     *logic.TipBuffer
}

func NewTunnel(tcpConn *logic.TcpConnection) *Tunnel {
	return &Tunnel{
		tcpConnection: tcpConn,
	}
}

func server(tunnel *Tunnel) {
	for {
		tipRequest := logic.NewTipBuffer()
		if err := tipRequest.ReadFrom(tunnel.tcpConnection); err != nil {
			continue
		}
		switch tipRequest.Opcode {
		case logic.OpcodeBind:

		}
	}
	//
	//d_tcpAddr, _ := net.ResolveTCPAddr("tcp4", "")
	//d_conn, err := net.DialTCP("tcp", nil, d_tcpAddr)
	//if err != nil {
	//	fmt.Println(err)
	//	s_conn.Write([]byte("can't connect "))
	//	s_conn.Close()
	//
	//}
	//go io.Copy(s_conn, d_conn)
	//go io.Copy(d_conn, s_conn)
	//
	//failed:
	//tcpConn.Close()
	//return
}
