package datalink

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go"
	"github.com/stitchcula/bacnet-go/layers"
	"net"
	"syscall"
)

const (
	BIPMaxAPDU     = 1476
	BIPMaxNPDU     = 1 + 1 + 2 + 1 + 7 + 2 + 1 + 7 + 1 + 1 + 2
	BIPMaxPDU      = BIPMaxAPDU + BIPMaxNPDU
	BIPMaxMPDU     = 1 + 1 + 2 + BIPMaxPDU
	BIPDefaultPort = 0xBAC0
)

type BIPConn struct {
	net.PacketConn
	msk       *net.IPNet
	laddr     *bacnet.Addr
	broadcast *bacnet.Addr
}

func NewBIPConn(ifn string) (c *BIPConn, err error) {
	h, _, err := net.SplitHostPort(ifn)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(h)
	if ip == nil || ip.To4() == nil || ip.IsUnspecified() {
		return nil, fmt.Errorf("invalid IPv4 %s", h)
	}

	c = &BIPConn{}
	c.PacketConn, err = net.ListenPacket("udp", ifn)
	if err != nil {
		return nil, err
	}

	laddr := c.PacketConn.LocalAddr().(*net.UDPAddr)
	c.laddr = bacnet.ResolveUDPAddr(laddr)

	if err = c.resolveBroadcast(laddr); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *BIPConn) WriteTo(apdu []byte, addr net.Addr) (n int, err error) {
	dst, ok := addr.(*bacnet.Addr)
	if !ok {
		return 0, &net.OpError{Op: "write", Net: addr.Network(), Source: c.laddr, Addr: addr, Err: syscall.EINVAL}
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
		ProtocolVersion:       layers.NPDUProtocolVersion,
		Dst:                   dst,
		Src:                   c.laddr,
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

func (c *BIPConn) ReadFrom(apdu []byte) (n int, addr net.Addr, err error) {
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

// resolveBroadcast resolve the broadcast IP to c.broadcast
func (c *BIPConn) resolveBroadcast(laddr *net.UDPAddr) error {
	uni, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for i := range uni {
		ip, msk, err := net.ParseCIDR(uni[i].String())
		if err != nil || ip.To4() == nil || !laddr.IP.Equal(ip) {
			continue
		}
		c.msk = msk
		broadcast := net.IP(make([]byte, 4))
		for i := range broadcast {
			broadcast[i] = msk.IP[i] | ^msk.Mask[i]
		}
		c.broadcast = bacnet.ResolveUDPAddr(&net.UDPAddr{
			IP:   broadcast,
			Port: BIPDefaultPort,
		})
		c.broadcast.SetBroadcast(true)
	}

	if c.broadcast == nil {
		return fmt.Errorf("can not bind network interface")
	}

	return nil
}
