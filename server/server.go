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
		fmt.Println("execCmd", tipRequest.DestIpToString(), tipRequest.DestPortToString(), tipRequest.Data)
		switch tipRequest.Opcode {
		case logic.OpcodeTransmit:
			if tunnel.connectedDest == false {
				var err error
				tunnel.destTcpConnection, err = logic.NewTcpConn(tipRequest.DestIpToString(), tipRequest.DestPortToString())
				if err != nil {
					fmt.Println("destTcpConnection error ", err)
					tunnel.quitSignal <- struct{}{}
					return
				}
				tunnel.connectedDest = true
				go tunnel.serveRead()
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
		fmt.Println("serveRead", string(buff))

		tip := logic.NewTipBuffer()
		transmitStream := tip.DataToTransmitStream(buff)
		_, err = tunnel.tcpConnection.Write(transmitStream)
		if err != nil {
			tunnel.quitSignal <- struct{}{}
			return
		}
	}
}
