package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
	"tcp_tunnel/logic"
	"tcp_tunnel/config"
)

var (
	localIp, localPort, remoteIp, remotePort string
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&localIp, "i", "127.0.0.1", "local ip address")
	flag.StringVar(&localPort, "p", "8888", "local port address")
	flag.StringVar(&remoteIp, "ri", "127.0.0.1", "remote ip address")
	flag.StringVar(&remotePort, "rp", "6379", "remote port address")
	flag.Parse()
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", localIp, localPort))
	if err != nil {
		panic("resolvetcpaddr failed.")
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic("listen ip is nil")
	}
	defer listener.Close()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic("accept error.")
		}
		go transmit(conn)
	}
}

func transmit(localConn *net.TCPConn) {
	serverTcpAddr, err := net.ResolveTCPAddr("tcp4", config.ServerIp+":"+config.ServerPort)
	if err != nil {
		panic("resolve remote tcp addr failed.")
		return
	}
	serverTcpConn, err := net.DialTCP("tcp", nil, serverTcpAddr)
	if err != nil {
		panic("dial remote tcp failed.")
		return
	}
	serverConn := logic.NewTcpContection(serverTcpConn)
	tip := logic.NewTipBuffer()
	bindStream := tip.BindStream(remoteIp, remotePort)
	_, err = serverConn.Write(bindStream)
	if err != nil {
		fmt.Print(err.Error())
	}
	go writeToServer(localConn, serverConn)
	go readServer(localConn, serverConn)
}

func writeToServer(localConn *net.TCPConn, serverConn *logic.TcpConnection) {
	for {
		data, err := ioutil.ReadAll(localConn)
		if err != nil {
			errors.New("read data error.")
		}
		tip := logic.NewTipBuffer()
		transmitStream := tip.TransmitStream(remoteIp, remotePort, data)
		_, err = serverConn.Write(transmitStream)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}

func readServer(localConn *net.TCPConn, serverConn *logic.TcpConnection) {
	for {
		if err := serverConn.ReadProtoBuffer(); err != nil {
			fmt.Println("ReadProtoBuffer", err.Error())
			continue;
		}
		opcode, err := serverConn.ReadOpcode()
		if err != nil {
			fmt.Println("ReadOpcode", err.Error())
			continue;
		}
		if opcode == logic.OpcodeBindAck {
			continue;
		}
		data, err := serverConn.ReadData()
		if err != nil {
			fmt.Println("ReadData", err.Error())
			continue;
		}
		_, err = localConn.Write(data)
		if err != nil {
			fmt.Println("write", err.Error())
		}
	}
}
