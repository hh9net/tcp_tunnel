package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"runtime"
	"tcp_tunnel/logic"
)

var (
	localIp, localPort, serverIP, serverPort, remoteIp, remotePort string
	quitSignal                                                     = make(chan struct{}, 1)
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&localIp, "i", "127.0.0.1", "local ip address")
	flag.StringVar(&localPort, "p", "8888", "local port address")
	flag.StringVar(&serverIP, "si", "127.0.0.1", "server ip address")
	flag.StringVar(&serverPort, "sp", "8787", "server port address")
	flag.StringVar(&remoteIp, "ri", "127.0.0.1", "remote ip address")
	flag.StringVar(&remotePort, "rp", "6379", "remote port address")
	flag.Parse()
}

func main() {
	listener, err := logic.NewTcpListener(localIp, localPort)
	if err != nil {
		panic("listen ip is nil")
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic("accept error.")
		}
		go transmit(conn)
	}
}

func transmit(localConn *net.TCPConn) {
	serverTcpConn, err := logic.NewTcpConn(serverIP, serverPort)
	if err != nil {
		panic("connect server failed.")
		return
	}
	serverConn := logic.NewTcpContection(serverTcpConn)
	go writeToServer(localConn, serverConn)
	go readServer(localConn, serverConn)
}

func writeToServer(localConn *net.TCPConn, serverConn *logic.TcpConnection) {
	for {
		buff := make([]byte, logic.ReadBuffLen)
		_, err := localConn.Read(buff)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Println("localConn read error", err.Error())
			continue
		}
		if len(buff) <= 0 {
			continue
		}
		tip := logic.NewTipBuffer()
		transmitStream := tip.TransmitStream(remoteIp, remotePort, buff)
		_, err = serverConn.Write(transmitStream)
		if err != nil {
			fmt.Print("serverConn write error", err.Error())
			quitSignal <- struct{}{}
		}
		fmt.Print("write data", string(buff), transmitStream, remoteIp, tip.Opcode, tip.DestIp, tip.DestPort, tip.DataLen, tip.Data)
		fmt.Printf("writeToServer: %#v", tip)
		select {
		case <-quitSignal:
			return
		}
	}
}

func readServer(localConn *net.TCPConn, serverConn *logic.TcpConnection) {
	for {
		if err := serverConn.ReadProtoBuffer(); err != nil {
			fmt.Println("ReadProtoBuffer", err.Error())
			continue
		}
		dataLen, err := serverConn.ReadDataLen()
		if err != nil {
			fmt.Println("ReadOpcode", err.Error())
			continue
		}
		buffLen := int(dataLen)
		data, err := serverConn.ReadData(buffLen)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Println("ReadData", err.Error())
			quitSignal <- struct{}{}
			return
		}
		_, err = localConn.Write(data)
		if err != nil {
			fmt.Println("write", err.Error())
		}
		fmt.Printf("readServer: %#v", data)
		select {
		case <-quitSignal:
			return
		}
	}
}
