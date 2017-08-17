package server

import (
	"fmt"
	"net"
	"tcp_tunnel/config"
	"tcp_tunnel/logic"
)

func TunnelListen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", config.ServerIp, config.ServerPort))
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic("listen ip is nil")
	}
	for {
		tcpConntection, err := logic.Accept(tcpListener)
		if err != nil {
			continue
		}
		go server(NewTunnel(tcpConntection))
	}
}

type Tunnel struct {
	bound             bool
	tcpConnection     *logic.TcpConnection
	tipBuffer         *logic.TipBuffer
	destTcpConnection *net.TCPConn
	quitSignal        chan struct{}
}

func NewTunnel(tcpConn *logic.TcpConnection) *Tunnel {
	return &Tunnel{
		bound:         false,
		tcpConnection: tcpConn,
		quitSignal:    make(chan struct{}, 1),
	}
}

func server(tunnel *Tunnel) {
	go tunnel.execCmd()
	for {
		select {
		case <-tunnel.quitSignal:
			goto failed
		}
	}
failed:
	return
}

func (tunnel *Tunnel) execCmd() {
Exec:
	for {
		tipRequest := logic.NewTipBuffer()
		if err := tipRequest.ReadFrom(tunnel.tcpConnection); err != nil {
			continue
		}
		fmt.Printf("readform: %#v", tipRequest)
		switch tipRequest.Opcode {
		case logic.OpcodeBind:
			if tunnel.bound == true {
				continue Exec
			}
			tcpAddr, err := net.ResolveTCPAddr("tcp4", tipRequest.DestIpToString()+":"+tipRequest.DestPortToString())
			if err != nil {
				tunnel.quitSignal <- struct{}{}
				return
			}
			tunnel.destTcpConnection, err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				tunnel.quitSignal <- struct{}{}
				return
			}
			bindAckStream := tipRequest.StreamTip(logic.OpcodeBindAck)
			_, err = tunnel.tcpConnection.Write(bindAckStream)
			if err != nil {
				tunnel.quitSignal <- struct{}{}
				return
			}
			go tunnel.serveRead()
			tunnel.bound = true
		case logic.OpcodeTransmit:
			tunnel.destTcpConnection.Write(tipRequest.Data)
		}
	}
}

func (tunnel *Tunnel) serveRead() {
	for {
		buff := make([]byte, logic.ReadBuffLen)
		_, err := tunnel.destTcpConnection.Read(buff)
		if err != nil {
			tunnel.quitSignal <- struct{}{}
			return
		}
		_, err = tunnel.tcpConnection.Write(buff)
		if err != nil {
			tunnel.quitSignal <- struct{}{}
			return
		}
	}
}
