package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"runtime"
	"tcp_tunnel/config"
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
	flag.StringVar(&serverIP, "i", "127.0.0.1", "server ip address")
	flag.StringVar(&serverPort, "p", "8787", "server port address")
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
		if err != nil {
			errors.New("read data error.")
		}
		if len(buff) <= 0 {
			continue
		}
		tip := logic.NewTipBuffer()
		transmitStream := tip.TransmitStream(remoteIp, remotePort, buff)
		_, err = serverConn.Write(transmitStream)
		if err != nil {
			fmt.Print(err.Error())
		}
		fmt.Printf("writeToServer: %#v", transmitStream)
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
		opcode, err := serverConn.ReadOpcode()
		if err != nil {
			fmt.Println("ReadOpcode", err.Error())
			continue
		}
		if opcode == logic.OpcodeBindAck {
			continue
		}
		data, err := serverConn.ReadData()
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
	}
}
