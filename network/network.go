package network

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go"
	"github.com/stitchcula/bacnet-go/datalink"
	"github.com/stitchcula/bacnet-go/layers"
	"net"
	"syscall"
	"unicode"
)

const (
	BIPMaxAPDU     = 1476
	BIPMaxNPDU     = 1 + 1 + 2 + 1 + 7 + 2 + 1 + 7 + 1 + 1 + 2
	BIPMaxPDU      = BIPMaxAPDU + BIPMaxNPDU
	BIPMaxMPDU     = 1 + 1 + 2 + BIPMaxPDU
	BIPDefaultPort = 0xBAC0
)

type Network struct {
	datalink.DataLink

	done   chan struct{}
	recvCh chan struct{}
	sendCh chan struct{}
}

func ListenPacket(typ datalink.Type, ifn string) (net.PacketConn, error) {
	dl, err := datalink.NewDataLink(typ, ifn)
	if err != nil {
		return nil, err
	}

	c := &Network{
		DataLink: dl,
		done:     make(chan struct{}),
		recvCh:   make(chan struct{}),
		sendCh:   make(chan struct{}),
	}

	go c.sendLoop()
	go c.recvLoop()

	return c, nil
}

func (c *Network) sendLoop() {

}

func (c *Network) recvLoop() {
	for {
		select {
		case <-c.done:
			return
		case bf, ok:=<-c.recvCh:

		}
	}
}

func (c *Network) WriteTo(apdu []byte, addr net.Addr) (n int, err error) {
	dst, ok := addr.(*bacnet.Addr)
	if !ok {
		return 0, &net.OpError{Op: "write", Net: addr.Network(), Source:, Addr: addr, Err: syscall.EINVAL}
	}
	if dst == bacnet.BroadcastAddr {
		dst = c.broadcast
	}

	// link layer
	vlc := &layers.BVLC{
		Type: layers.BVLCTypeBIP,
	}
	if dst.IsBroadcast() || dst.IsSubBroadcast() {
		vlc.Function = layers.BVLCOriginalBroadcastNPDU
	} else {
		vlc.Function = layers.BVLCOriginalUnicastNPDU
	}

	// network layer
	npdu := &layers.NPDU{
		ProtocolVersion: layers.NPDUProtocolVersion,
		Dst:             dst,
		Src:             c.laddr,
		IsNetworkLayerMessage: false,
		ExpectingReply:        false,
		Priority:              layers.NPDUPriorityNormal,
		HopCount:              layers.NPDUDefaultHopCount,
	}

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{},
		vlc,
		npdu,
		gopacket.Payload(apdu),
	)

	return c.PacketConn.WriteTo(buf.Bytes(), dst.UDPAddr())
}

func (c *Network) ReadFrom(apdu []byte) (n int, addr net.Addr, err error) {
	data := make([]byte, 0, BIPMaxMPDU)
	for {
		n, addr, err = c.PacketConn.ReadFrom(data)
		if err != nil {
			return
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeBACnetVLC, gopacket.Default)
		layer := packet.Layer(layers.LayerTypeBACnetVLC)
		if layer == nil {
			continue
		}
		vlc, _ := layer.(*layers.BVLC)
		switch vlc.Function {
		case layers.BVLCResult:
			fmt.Printf("BVLC: Result Code=%d\r\n", binary.BigEndian.Uint16(vlc.LayerPayload()))
		case layers.BVLCWriteBroadcastDistTable:
			fmt.Println("BVLC: layers.BVLCWriteBroadcastDistributionTable")
		case layers.BVLCReadBroadcastDistTable:
			fmt.Println("BVLC: layers.BVLCReadBroadcastDistTable")
		case layers.BVLCReadBroadcastDistTableAck:
			fmt.Println("BVLC: layers.BVLCReadBroadcastDistTableAck")
		case layers.BVLCForwardedNPDU:

		}
	}
}
