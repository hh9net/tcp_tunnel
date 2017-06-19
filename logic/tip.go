package logic

import (
	"net"
	"strconv"
	"strings"
	"log"
	"errors"
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
		errors.New("read proto buffer error.")
	}
	t.Opcode, err = tcpConnection.ReadOpcode()
	if err != nil {
		errors.New("read opcode error.")
	}
	return nil
}

func (t *TipBuffer) StreamTip() []byte {
	switch t.Opcode {
	case OpcodeBind:
		buffer := make([]byte, 0, 1)
		buffer[0] = byte(OpcodeBindAck)
		return buffer
	}
	return []byte{}
}

func (t *TipBuffer) WriteTo(tcpConn *TcpConnection) error {
	buffer := t.StreamTip()
	_, err := tcpConn.Write(buffer)
	return err
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
