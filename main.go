package main

import (
    "fmt"
    "io"
    "net"
    "os"
)

func main() {
    l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ListenIp, ListenPort))
    if err != nil {
	fmt.Println(err, err.Error())
	os.Exit(0)
    }

    for {
	s_conn, err := l.Accept()
	if err != nil {
	    continue
	}

	d_tcpAddr, _ := net.ResolveTCPAddr("tcp4", "10.32.39.205:3306")
	d_conn, err := net.DialTCP("tcp", nil, d_tcpAddr)
	if err != nil {
	    fmt.Println(err)
	    s_conn.Write([]byte("can't connect 10.32.39.205:3306"))
	    s_conn.Close()
	    continue
	}
	go io.Copy(s_conn, d_conn)
	go io.Copy(d_conn, s_conn)
    }
}
