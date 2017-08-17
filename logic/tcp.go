package logic

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	ReadBuffLen            = 0xFFFFFF
	TcpProtoBufferLen      = 7
	ProtoOpcodeBufferLen   = 1
	ProtoDestIpBufferLen   = 4
	ProtoDestPortBufferLen = 2
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

func (tc *TcpConnection) ReadData() ([]byte, error) {
	buff := make([]byte, ReadBuffLen)
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
