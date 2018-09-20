package bacnet

import (
	"encoding/binary"
	"fmt"
	"net"
)

const BroadcastNetwork uint16 = 0xFFFF

var (
	BroadcastAddr = &Addr{}
)

type Addr struct {
	MacLen byte
	Mac    [7]byte
	Net    uint16
	Len    byte
	Adr    [7]byte
}

func (addr *Addr) Network() string {
	return "BACnet"
}

func (addr *Addr) String() string {
	return fmt.Sprintf("%x", addr.Adr[:])
}

func (addr *Addr) IsBroadcast() bool {
	if addr.Net == BroadcastNetwork || addr.MacLen == 0 {
		return true
	}
	return false
}

func (addr *Addr) SetBroadcast(b bool) {
	if b {
		addr.MacLen = 0
	} else {
		addr.MacLen = uint8(len(addr.Mac))
	}
}

func (addr *Addr) IsSubBroadcast() bool {
	if addr.Net > 0 && addr.Len == 0 {
		return true
	}
	return false
}

func (addr *Addr) IsUnicast() bool {
	if addr.MacLen == 6 {
		return true
	}
	return false
}

func (addr *Addr) UDPAddr() *net.UDPAddr {
	port := uint(addr.Mac[4])<<8 | uint(addr.Mac[5])

	return &net.UDPAddr{
		IP:   net.IP(addr.Mac[:4]),
		Port: int(port),
	}
}

func (addr *Addr) DecodeFromBytes(data []byte) (offset int) {
	addr.Net = binary.BigEndian.Uint16(data[:2])
	addr.Len = data[2]
	for i := byte(0); i < addr.Len; i++ {
		addr.Adr[i] = data[3+i]
	}
	return int(3 + addr.Len)
}

func (addr *Addr) Bytes() []byte {
	b := make([]byte, 3, 10)
	binary.BigEndian.PutUint16(b, addr.Net)
	b[2] = addr.Len
	b = append(b, addr.Adr[:addr.Len]...)
	return b
}

func ResolveUDPAddr(addr *net.UDPAddr) *Addr {
	a := Addr{
		MacLen: uint8(net.IPv4len + 2),
	}
	p := uint16(addr.Port)

	for i := 0; i < net.IPv4len; i++ {
		a.Mac[i] = addr.IP[i]
	}

	a.Mac[net.IPv4len+0] = byte(p >> 8)
	a.Mac[net.IPv4len+1] = byte(p & 0x00FF)

	return &a
}
