package server

import (
	"fmt"
	"io/ioutil"
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
		bound:   false,
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
	for {
		tipRequest := logic.NewTipBuffer()
		if err := tipRequest.ReadFrom(tunnel.tcpConnection); err != nil {
			continue
		}
		switch tipRequest.Opcode {
		case logic.OpcodeBind:
			tcpAddr, err := net.ResolveTCPAddr("tcp4", tipRequest.DestIpToString() + ":" + tipRequest.DestPortToString())
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
	if (tunnel.bound == true) {
		return
	}
	for {
		data, err := ioutil.ReadAll(tunnel.destTcpConnection)
		if err != nil {
			tunnel.quitSignal <- struct{}{}
			return
		}
		_, err = tunnel.tcpConnection.Write(data)
		if err != nil {
			tunnel.quitSignal <- struct{}{}
			return
		}
	}
}
