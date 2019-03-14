package datalink

import (
	"errors"
	"net"
)

type DataLink = net.PacketConn

type Type int

const (
	Ethernet Type = iota
	ARCnet
	MSTP
	BIP
)

func NewDataLink(typ Type, ifn string) (DataLink, error) {
	switch typ {
	case BIP:
		return ListenBIP(ifn)
	}

	return nil, errors.New("not support type")
}
