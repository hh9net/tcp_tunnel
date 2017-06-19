package logic

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
	"idgen/logic"
)

const (
	tcpProtoBufferLen      = 11
	protoOpcodeBufferLen   = 1
	protoDestIpBufferLen   = 8
	protoDestPortBufferLen = 2
)

type TcpConnection struct {
	*net.TCPConn
	protoBuffer []byte
}

func Accept(listener *net.Listener) (*TcpConnection, error) {
	tc, err := listener.Accept()
	if err != nil {
		nil, errors.New("listen error")
	}
	tcpConn := NewTcpContection(tc)
	return tcpConn, nil
}

func NewTcpContection(tcpConn *net.TCPConn) *TcpConnection {
	return &TcpConnection{TCPConn: tcpConn, protoBuffer: make([]byte, tcpProtoBufferLen)}
}

func (tc *TcpConnection) ReadProtoBuffer() error {
	left := tcpProtoBufferLen
	for left > 0 {
		n, err := tc.TCPConn.Read(tc.protoBuffer)
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
		return nil, errors.New("protoBuffer is nil")
	}
	opcode := uint8(tc.protoBuffer[0:protoOpcodeBufferLen])
	return opcode, nil
}

func (tc *TcpConnection) ReadDestIp() (uint64, error)  {
	if tc.protoBuffer == nil {
		return nil, errors.New("protoBuffer is nil")
	}
	destIp := binary.BigEndian.Uint64(tc.protoBuffer[protoOpcodeBufferLen : protoOpcodeBufferLen+protoDestIpBufferLen])
	return destIp, nil
}

func (tc *TcpConnection) ReadPort() (uint16, error) {
	if tc.protoBuffer == nil {
		return nil, errors.New("protoBuffer is nil")
	}
	start := protoOpcodeBufferLen + protoDestIpBufferLen
	end := protoOpcodeBufferLen + protoDestIpBufferLen + protoDestPortBufferLen
	destPort := binary.BigEndian.Uint64(tc.protoBuffer[start:end])
	return destPort, nil
}

func (tc *TcpConnection) ReadData() error {
	data, err := ioutil.ReadAll(tc.TCPConn)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (tc *TcpConnection) RemoteIP() string {
	tcpAddr := tc.RemoteAddr().(*net.TCPAddr)
	return tcpAddr.IP.String()
}
