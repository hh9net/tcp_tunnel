package logic

import (
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

type Tip struct {
	Opcode   uint8
	DestIp   uint64
	DestPort uint16
	Data     []byte
}

func NewTip(op int8) *Tip {
	return &Tip{Opcode: op}
}

func (t *Tip) StreamTip() []byte {
	switch t.Opcode {
	case OpcodeBind:
		buffer := make([]byte, 0, 1)
		buffer[0] = byte(OpcodeBindAck)
		return buffer
	}
	return []byte{}
}

func (t *Tip) WriteTo(tcpConn *TcpConnection) error {
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
