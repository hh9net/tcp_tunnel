package server

import (
	"fmt"
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
	defer tunnel.tcpConnection.Close()
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
		fmt.Println("execCmd", tipRequest.Opcode, tipRequest.DestIpToString(), tipRequest.DestPortToString(), tipRequest.Data, tunnel.connectedDest)
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
				defer tunnel.destTcpConnection.Close()
				tunnel.connectedDest = true
				go tunnel.serveRead()
			}
			fmt.Println("exeeeeeeee", string(tipRequest.Data), len(tipRequest.Data))
			_, err := tunnel.destTcpConnection.Write(tipRequest.Data)
			if err != nil {
				fmt.Println("xxxxxxxxxxxxxxxxxx", err)
				tunnel.quitSignal <- struct{}{}
				return
			}
		}
	}
}

func (tunnel *Tunnel) serveRead() {
	for {
		buff, err := logic.ChunkRead(tunnel.destTcpConnection)
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
