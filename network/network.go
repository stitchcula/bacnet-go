package network

import (
	"github.com/stitchcula/bacnet-go/datalink"
	"net"
)

type Network struct {
	datalink.DataLink
}

func Dial(typ datalink.Type, ifn string) (net.PacketConn, error) {

}

func (c *Network) ReadFrom(p []byte) (n int, addr net.Addr, err error) {

}

func (c *Network) WriteTo(p []byte, addr net.Addr) (n int, err error) {

}
