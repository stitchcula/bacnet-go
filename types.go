package bacnet

type Addr struct {
	MacLen byte
	Mac    [7]byte
	Net    byte
	Len    byte
	Adr    [7]byte
}

func (addr *Addr) Network() string {
	return "BACnet"
}

func (addr *Addr) String() string {
	return string(addr.Adr[:])
}
