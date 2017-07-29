package server

import (
	"fmt"
	"io"
	"net"
	"tcp_tunnel/config"
	"tcp_tunnel/logic"
	"io/ioutil"
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
	tcpConnection     *logic.TcpConnection
	tipBuffer         *logic.TipBuffer
	destTcpConnection *net.TCPConn
	quitSignal        chan bool
}

func NewTunnel(tcpConn *logic.TcpConnection) *Tunnel {
	return &Tunnel{
		tcpConnection: tcpConn,
		quitSignal: make(chan bool, 1),
	}
}

func server(tunnel *Tunnel) {
	go tunnel.execCmd()
	go tunnel.serveRead()
	for {
		select {
		case <-tunnel.quitSignal:
			goto failed
		}
	}
	failed:
	return
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

func (tunnel *Tunnel) execCmd() {
	for {
		tipRequest := logic.NewTipBuffer()
		if err := tipRequest.ReadFrom(tunnel.tcpConnection); err != nil {
			continue
		}
		switch tipRequest.Opcode {
		case logic.OpcodeBind:
			tcpAddr, err := net.ResolveTCPAddr("tcp4", tipRequest.DestIp + ":" + tipRequest.DestPort)
			if err != nil {
				tunnel.quitSignal <- true
				return
			}
			tunnel.destTcpConnection, err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				tunnel.quitSignal <- true
				return
			}
			bindAckStream := tipRequest.StreamTip(logic.OpcodeBindAck)
			_, err := tunnel.tcpConnection.Write(bindAckStream)
			if err != nil {
				tunnel.quitSignal <- true
				return
			}
		case logic.OpcodeTransmit:
			tunnel.destTcpConnection.Write(tipRequest.Data)
		}
	}
}

func (tunnel *Tunnel) serveRead() {
	for {
		data, err := ioutil.ReadAll(tunnel.destTcpConnection)
		if err != nil {
			tunnel.quitSignal <- true
			return
		}
		_, err = tunnel.tcpConnection.Write(data)
		if err != nil {
			tunnel.quitSignal <- true
			return
		}
	}
}

