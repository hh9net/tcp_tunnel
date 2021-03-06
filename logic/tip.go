package logic

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	DestIp   uint32
	DestPort uint16
	DataLen  uint32
	Data     []byte
}

func NewTipBuffer() *TipBuffer {
	return &TipBuffer{}
}

func (t *TipBuffer) DestIpToString() string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, t.DestIp)
	return ip.String()
}

func (t *TipBuffer) DestPortToString() string {
	port := strconv.Itoa(int(t.DestPort))
	return port
}

func (t *TipBuffer) ReadFrom(tcpConnection *TcpConnection) error {
	if err := tcpConnection.ReadProtoBuffer(); err != nil {
		return errors.New("read proto buffer error.")
	}
	var err error
	t.Opcode, err = tcpConnection.ReadOpcode()
	fmt.Println("t.Opcode", t.Opcode, OpcodeTransmit)
	if err != nil {
		return errors.New("read opcode error.")
	}
	switch t.Opcode {
	case OpcodeTransmit:
		t.DestIp, err = tcpConnection.ReadDestIp()
		fmt.Println("t.DestIp", t.DestIp)
		if err != nil {
			return errors.New("read destIP error.")
		}
		t.DestPort, err = tcpConnection.ReadDestPort()
		fmt.Println("t.DestPort", t.DestPort)
		if err != nil {
			return errors.New("read destPort error.")
		}
		t.DataLen, err = tcpConnection.ReadDataLen()
		fmt.Println("t.DataLen", t.DataLen)
		if err != nil {
			return errors.New("read data len error.")
		}
		buffLen := int(t.DataLen)
		t.Data, err = tcpConnection.ReadData(buffLen)
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

func (t *TipBuffer) BindStream(destIp, destPort string) []byte {
	t.Opcode = OpcodeBind
	ip := net.ParseIP(destIp)
	port, _ := strconv.ParseUint(destPort, 10, 16)
	t.DestIp = binary.BigEndian.Uint32(ip)
	t.DestPort = uint16(port)

	buff := make([]byte, ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen)
	buff[ProtoOpcodeBufferLen] = t.Opcode
	binary.BigEndian.PutUint32(buff[ProtoOpcodeBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen], t.DestIp)
	binary.BigEndian.PutUint16(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen:TcpProtoBufferLen], t.DestPort)
	return buff
}

func (t *TipBuffer) TransmitStream(destIp, destPort string, data []byte) []byte {
	t.Opcode = OpcodeTransmit
	ip := []byte(net.ParseIP(destIp).To4())
	port, _ := strconv.ParseUint(destPort, 10, 16)
	t.DestIp = binary.BigEndian.Uint32(ip)
	t.DestPort = uint16(port)
	t.DataLen = uint32(len(data))
	fmt.Println("TransmitStream*******", destIp, ip, t.DestIp)

	buff := make([]byte, TcpProtoBufferLen+len(data))
	buff[ProtoOpcodeBufferLen-1] = t.Opcode
	fmt.Println("TransmitStream-----", t.Opcode, buff)
	binary.BigEndian.PutUint32(buff[ProtoOpcodeBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen], t.DestIp)
	fmt.Println("TransmitStream+++++", t.DestIp, buff)
	binary.BigEndian.PutUint16(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen], t.DestPort)
	fmt.Println("TransmitStream-----", t.DestPort, buff)
	binary.BigEndian.PutUint32(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen:TcpProtoBufferLen], t.DataLen)
	fmt.Println("TransmitStream+++++", t.DataLen, buff)
	copy(buff[TcpProtoBufferLen:], data)
	fmt.Println("TransmitStream=====", data, buff)
	return buff
}

func (t *TipBuffer) DataToTransmitStream(data []byte) []byte {
	t.Opcode = OpcodeTransmit
	t.DataLen = uint32(len(data))
	fmt.Println("TransmitStream*******", t.DestIp, t.DestPort)

	buff := make([]byte, TcpProtoBufferLen+len(data))
	buff[ProtoOpcodeBufferLen-1] = t.Opcode
	fmt.Println("TransmitStream-----", t.Opcode, buff)
	binary.BigEndian.PutUint32(buff[ProtoOpcodeBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen], t.DestIp)
	fmt.Println("TransmitStream+++++", t.DestIp, buff)
	binary.BigEndian.PutUint16(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen:ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen], t.DestPort)
	fmt.Println("TransmitStream-----", t.DestPort, buff)
	binary.BigEndian.PutUint32(buff[ProtoOpcodeBufferLen+ProtoDestIpBufferLen+ProtoDestPortBufferLen:TcpProtoBufferLen], t.DataLen)
	fmt.Println("TransmitStream+++++", t.DataLen, buff)
	copy(buff[TcpProtoBufferLen:], data)
	fmt.Println("TransmitStream=====", data, buff)
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
