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
		buff, err := logic.ChunkRead(localConn)
		if err != nil {
			fmt.Println("localConn read error", err.Error())
			continue
		}
		if len(buff) <= 0 {
			continue
		}
		fmt.Println("xxxxxxxxxxx", string(buff))
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

// todo 怀疑在读取redis-cli没有完全读完，写到redis-server端时，而且还要查下，在读取时的make buff，不要初始化为0，不然读取数据写入到buff里有问题，
func readServer(localConn *net.TCPConn, serverConn *logic.TcpConnection) {
	for {
		fmt.Println("readServer")
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
		fmt.Println("111111111111", buffLen, string(data), err)
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
