package logic

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	_ = iota
	OpcodeBind
	OpcodeBindAck
	OpcodeTransmit
)

type TipBuffer struct {
	Opcode   uint8
	DestIp   uint64
	DestPort uint16
	Data     []byte
}

func NewTipBuffer() *TipBuffer {
	return &TipBuffer{}
}

func (t *TipBuffer) ReadFrom(tcpConnection *TcpConnection) error {
	if err := tcpConnection.ReadProtoBuffer(); err != nil {
		return errors.New("read proto buffer error.")
	}
	t.Opcode, err = tcpConnection.ReadOpcode()
	if err != nil {
		return errors.New("read opcode error.")
	}
	switch t.Opcode {
	case OpcodeBind:
		t.DestIp, err = tcpConnection.ReadDestIp()
		if err != nil {
			return errors.New("read destIP error.")
		}
		t.DestPort, err = tcpConnection.ReadDestPort()
		if err != nil {
			return errors.New("read destPort error.")
		}
	case OpcodeTransmit:
		t.Data, err = tcpConnection.ReadData()
		if err != nil {
			return errors.New("read data error.")
		}
	}
	return nil
}

func (t *TipBuffer) StreamTip(opcode int) []byte {
	switch opcode {
	case OpcodeBindAck:
		buffer := make([]byte, 0, 1)
		buffer[0] = byte(OpcodeBindAck)
		return buffer
	}
	return []byte{}
}

func (t *TipBuffer) TransmitStream(ip, port string, data []byte) []byte {
	t.Opcode = OpcodeTransmit
	t.DestIp = ip
	t.DestPort = port
	t.Data = data

	buff := make([]byte, ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen+len(data))
	buff[0:ProtoOpcodeBufferLen] = t.Opcode
	binary.BigEndian.PutUint64(buff[ProtoOpcodeBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen], t.DestIp)
	binary.BigEndian.PutUint16(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen:TcpProtoBufferLen], t.DestPort)
	copy(buff[TcpProtoBufferLen:], data)
	return buff
}

func IP4ToInt64(ipv4 net.IP) int64 {
	bits := strings.Split(ipv4.String(), ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum
}
