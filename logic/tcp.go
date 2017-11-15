package logic

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	ReadBuffLen            = 0xFFFFFF
	TcpProtoBufferLen      = 11
	ProtoOpcodeBufferLen   = 1
	ProtoDestIpBufferLen   = 4
	ProtoDestPortBufferLen = 2
	ProtoDataLen           = 4
)

type TcpConnection struct {
	*net.TCPConn
	protoBuffer []byte
}

func Accept(tcpListner *net.TCPListener) (*TcpConnection, error) {
	conn, err := tcpListner.AcceptTCP()
	if err != nil {
		return nil, errors.New("listen error")
	}
	tcpConn := NewTcpContection(conn)
	return tcpConn, nil
}

func NewTcpContection(tcpConn *net.TCPConn) *TcpConnection {
	return &TcpConnection{TCPConn: tcpConn, protoBuffer: make([]byte, TcpProtoBufferLen)}
}

func (tc *TcpConnection) ReadProtoBuffer() error {
	left := TcpProtoBufferLen
	for left > 0 {
		n, err := tc.TCPConn.Read(tc.protoBuffer)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n > 0 {
			left -= n
		}
	}
	return nil
}

func (tc *TcpConnection) ReadOpcode() (uint8, error) {
	if tc.protoBuffer == nil {
		return uint8(0), errors.New("protoBuffer is nil")
	}
	opcode := uint8(tc.protoBuffer[ProtoOpcodeBufferLen])
	return opcode, nil
}

func (tc *TcpConnection) ReadDestIp() (uint32, error) {
	if tc.protoBuffer == nil {
		return uint32(0), errors.New("protoBuffer is nil")
	}
	destIp := binary.BigEndian.Uint32(tc.protoBuffer[ProtoOpcodeBufferLen : ProtoOpcodeBufferLen+ProtoDestIpBufferLen])
	return destIp, nil
}

func (tc *TcpConnection) ReadDestPort() (uint16, error) {
	if tc.protoBuffer == nil {
		return uint16(0), errors.New("protoBuffer is nil")
	}
	start := ProtoOpcodeBufferLen + ProtoDestIpBufferLen
	end := ProtoOpcodeBufferLen + ProtoDestIpBufferLen + ProtoDestPortBufferLen
	destPort := binary.BigEndian.Uint16(tc.protoBuffer[start:end])
	return destPort, nil
}

func (tc *TcpConnection) ReadDataLen() (uint32, error) {
	if tc.protoBuffer == nil {
		return uint32(0), errors.New("protoBuffer is nil")
	}
	start := ProtoOpcodeBufferLen + ProtoDestIpBufferLen + ProtoDestPortBufferLen
	end := ProtoOpcodeBufferLen + ProtoDestIpBufferLen + ProtoDestPortBufferLen + ProtoDataLen
	dataLen := binary.BigEndian.Uint32(tc.protoBuffer[start:end])
	return dataLen, nil
}

func (tc *TcpConnection) ReadData(buffLen int) ([]byte, error) {
	buff := make([]byte, buffLen)
	_, err := tc.TCPConn.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (tc *TcpConnection) RemoteIP() string {
	tcpAddr := tc.RemoteAddr().(*net.TCPAddr)
	return tcpAddr.IP.String()
}

func NewTcpListener(ip, port string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return tcpListener, nil
}

func NewTcpConn(ip, port string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		panic("resolve remote tcp addr failed.")
		return nil, errors.New("resolve failed.")
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic("dial remote tcp failed.")
		return nil, errors.New("dial failed.")
	}
	return tcpConn, nil
}
