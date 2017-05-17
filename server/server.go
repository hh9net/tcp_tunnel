package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"tcp_tunnel/config"
	"tcp_tunnel/logic"
)

type Server struct{}

func StartServer() {
	s := new(Server)
	s.Server()
}

func (srv *Server) Server() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ListenIp, config.ListenPort))
	if err != nil {
		panic("listen ip is nil")
	}
	for {
		tc, err := listener.Accept()
		if err != nil {
			continue
		}
		tcpConn, err := logic.NewTcpContection(tc)

		d_tcpAddr, _ := net.ResolveTCPAddr("tcp4", "")
		d_conn, err := net.DialTCP("tcp", nil, d_tcpAddr)
		if err != nil {
			fmt.Println(err)
			s_conn.Write([]byte("can't connect "))
			s_conn.Close()
			continue
		}
		go io.Copy(s_conn, d_conn)
		go io.Copy(d_conn, s_conn)
	}
}
