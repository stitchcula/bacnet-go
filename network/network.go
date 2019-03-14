package network

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go"
	"github.com/stitchcula/bacnet-go/datalink"
	"github.com/stitchcula/bacnet-go/layers"
	"net"
	"syscall"
)

type frame struct {
	addr    net.Addr
	payload []byte
}

type Network struct {
	datalink.DataLink

	done   chan struct{}
	recvCh chan frame
	errCh  chan error
}

func ListenPacket(typ datalink.Type, ifn string) (net.PacketConn, error) {
	dl, err := datalink.NewDataLink(typ, ifn)
	if err != nil {
		return nil, err
	}

	c := &Network{
		DataLink: dl,
		done:     make(chan struct{}),
		recvCh:   make(chan frame),
		errCh:    make(chan error),
	}

	go c.recvLoop()

	return c, nil
}

func (c *Network) recvLoop() {
	data := make([]byte, 0, layers.MaxAPDU)
	for {
		select {
		case <-c.done:
			return
		default:
		}

		n, addr, err := c.DataLink.ReadFrom(data)
		if err != nil {
			c.errCh <- err
			continue
		}

		fmt.Println(data[0:n], addr)
	}
}

func (c *Network) WriteTo(apdu []byte, addr net.Addr) (n int, err error) {
	dst, ok := addr.(*bacnet.Addr)
	if !ok {
		return 0, &net.OpError{Op: "write", Net: addr.Network(), Source: c.DataLink.LocalAddr(), Addr: addr, Err: syscall.EINVAL}
	}

	npdu := &layers.NPDU{
		ProtocolVersion:       layers.NPDUProtocolVersion,
		Dst:                   dst,
		Src:                   nil,
		IsNetworkLayerMessage: false,
		ExpectingReply:        false,
		Priority:              layers.NPDUPriorityNormal,
		HopCount:              layers.NPDUDefaultHopCount,
	}

	bf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(bf, gopacket.SerializeOptions{},
		npdu,
		gopacket.Payload(apdu),
	)

	return c.DataLink.WriteTo(bf.Bytes(), addr)
}

func (c *Network) ReadFrom(apdu []byte) (n int, addr net.Addr, err error) {
	select {
	case <-c.done:
		return -1, nil, &net.OpError{Op: "read", Net: addr.Network(), Source: c.DataLink.LocalAddr(), Addr: addr, Err: syscall.EPIPE}
	case err := <-c.errCh:
		return -1, nil, err
	case fr := <-c.recvCh:
		copy(apdu, fr.payload)
		return len(fr.payload), fr.addr, nil
	}
}

func (c *Network) Close() error {
	close(c.done)
	err := c.DataLink.Close()
	close(c.recvCh)
	close(c.errCh)
	return err
}
