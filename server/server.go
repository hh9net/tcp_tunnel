package server

import (
	"fmt"
	"io"
	"net"
	"tcp_tunnel/config"
	"tcp_tunnel/logic"
)

func TunnelListen(conf *config.Config) {
	tcpListener, err := logic.NewTcpListener(conf.ServIp, conf.ServPort)
	if err != nil {
		panic(err)
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
	connectedDest     bool
	tcpConnection     *logic.TcpConnection
	tipBuffer         *logic.TipBuffer
	destTcpConnection *net.TCPConn
	quitSignal        chan struct{}
}

func NewTunnel(tcpConn *logic.TcpConnection) *Tunnel {
	return &Tunnel{
		connectedDest: false,
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
		fmt.Printf("readform: %#v", tipRequest)
		switch tipRequest.Opcode {
		case logic.OpcodeTransmit:
			if tunnel.connectedDest == false {
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
				tunnel.connectedDest = true
				tunnel.serveRead()
			}
			tunnel.destTcpConnection.Write(tipRequest.Data)
		}
	}
}

func (tunnel *Tunnel) serveRead() {
	for {
		buff := make([]byte, logic.ReadBuffLen)
		_, err := tunnel.destTcpConnection.Read(buff)
		if err == io.EOF {
			continue
		}
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
