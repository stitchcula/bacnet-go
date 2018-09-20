package bacnet

import (
	"encoding/binary"
	"net"
)

type Addr struct {
	MacLen byte
	Mac    [7]byte
	Net    uint16
	Len    byte
	Adr    [7]byte
}

func NewAddrUDP() *Addr {
	return &Addr{}
}

func NewAddrBytes(data []byte) *Addr {
	addr := &Addr{}
	addr.DecodeFromBytes(data)
	return addr
}

func (addr *Addr) Network() string {
	return "BACnet"
}

func (addr *Addr) String() string {
	return net.IP(addr.Mac[:4]).String()
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
