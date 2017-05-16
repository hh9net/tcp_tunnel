package logic

import (
    "strings"
    "strconv"
    "net"
)

const (
    _ = iota
    OpBind
    OpContent
)

type Tip struct {
    Op       int8
    DestIp   int64
    DestPort int16
    Data     []byte
}

func NewTip(op int8) *Tip {
    return &Tip{Op: op}
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
