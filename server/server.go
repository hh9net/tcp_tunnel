package server

import (
	"fmt"
	"io"
	"net"
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
		tcpConn := logic.NewTcpContection(tc)
		go srv.dispatch(tcpConn)
	}
}

func (srv *Server) dispatch(tcpConn *logic.TcpConnection) {
	if err := tcpConn.ReadProtoBuffer(); err != nil {
		goto failed
	}
	opcode, err := tcpConn.ReadOpcode()
	if err != nil {
		goto failed
	}
	tip := logic.NewTip(opcode)
	if

	d_tcpAddr, _ := net.ResolveTCPAddr("tcp4", "")
	d_conn, err := net.DialTCP("tcp", nil, d_tcpAddr)
	if err != nil {
		fmt.Println(err)
		s_conn.Write([]byte("can't connect "))
		s_conn.Close()

	}
	go io.Copy(s_conn, d_conn)
	go io.Copy(d_conn, s_conn)

	failed:
	tcpConn.Close()
	return
}
