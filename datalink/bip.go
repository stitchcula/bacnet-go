package datalink

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go"
	"github.com/stitchcula/bacnet-go/layers"
	"net"
	"syscall"
	"time"
)

const (
	BIPMaxMPDU     = 1 + 1 + 2 + layers.MaxPDU
	BIPDefaultPort = 0xBAC0
)

type frame struct {
	addr    net.Addr
	payload []byte
}

type BIPConn struct {
	net.PacketConn

	laddr     *bacnet.Addr
	broadcast *bacnet.Addr

	sess map[string][]byte

	done   chan struct{}
	recvCh chan frame
	errCh  chan error
}

func ListenBIP(ifn string) (c *BIPConn, err error) {
	h, _, err := net.SplitHostPort(ifn)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(h)
	if ip == nil || ip.To4() == nil || ip.IsUnspecified() {
		return nil, fmt.Errorf("invalid IPv4 %s", h)
	}

	conn, err := net.ListenPacket("udp", ifn)
	if err != nil {
		return nil, err
	}
	laddr := conn.LocalAddr().(*net.UDPAddr)

	c = &BIPConn{
		PacketConn: conn,
		laddr:      bacnet.ResolveUDPAddr(laddr),
		sess:       make(map[string][]byte),
		done:       make(chan struct{}),
		recvCh:     make(chan frame),
		errCh:      make(chan error),
	}
	if err = c.resolveBroadcast(laddr); err != nil {
		return nil, err
	}

	go c.recvLoop()

	return c, nil
}

func (c *BIPConn) recvLoop() {
	data := make([]byte, 0, BIPMaxMPDU)
	for {
		select {
		case <-c.done:
			return
		default:
		}

		n, addr, err := c.PacketConn.ReadFrom(data)
		if err != nil {
			c.errCh <- err
			continue
		} else if n < 1 {
			time.Sleep(time.Millisecond)
			continue
		}

		if sess, ok := c.sess[addr.String()]; ok {
			data = append(sess, data[:n]...)
			c.sess[addr.String()] = sess[:0]
		}

		if len(data) < 4 {
			c.sess[addr.String()] = data
			continue
		} else if data[0] != layers.BVLCTypeBIP {
			continue
		} else if length := binary.BigEndian.Uint16(data[2:4]); uint16(len(data)) < length {
			c.sess[addr.String()] = data
			continue
		} else if uint16(len(data)) > length {
			c.sess[addr.String()] = data[length:]
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeBACnetVLC, gopacket.Default)
		layer := packet.Layer(layers.LayerTypeBACnetVLC)
		if layer == nil {
			continue
		}
		bvlc, _ := layer.(*layers.BVLC)
		fmt.Println(bvlc.Payload)
		switch bvlc.Function {
		case layers.BVLCResult:
			fmt.Printf("BVLC: Result Code=%d\r\n", binary.BigEndian.Uint16(bvlc.LayerPayload()))
		case layers.BVLCWriteBroadcastDistTable:
			fmt.Println("BVLC: layers.BVLCWriteBroadcastDistributionTable")
		case layers.BVLCReadBroadcastDistTable:
			fmt.Println("BVLC: layers.BVLCReadBroadcastDistTable")
		case layers.BVLCReadBroadcastDistTableAck:
			fmt.Println("BVLC: layers.BVLCReadBroadcastDistTableAck")
		case layers.BVLCForwardedNPDU:
			fmt.Println("BVLC: layers.BVLCReadBroadcastDistTableAck")
		}
	}
}

func (c *BIPConn) WriteTo(npdu []byte, addr net.Addr) (n int, err error) {
	dst, ok := addr.(*bacnet.Addr)
	if !ok {
		return -1, &net.OpError{Op: "write", Net: addr.Network(), Source: c.laddr, Addr: addr, Err: syscall.EINVAL}
	}

	if dst.IsBroadcast() {
		dst.MacLen = c.broadcast.MacLen
		dst.Mac = c.broadcast.Mac
	}

	bvlc := &layers.BVLC{
		Type: layers.BVLCTypeBIP,
	}
	if dst.IsBroadcast() || dst.IsSubBroadcast() {
		bvlc.Function = layers.BVLCOriginalBroadcastNPDU
	} else {
		bvlc.Function = layers.BVLCOriginalUnicastNPDU
	}

	bf := gopacket.NewSerializeBuffer()
	if err = gopacket.SerializeLayers(bf, gopacket.SerializeOptions{},
		bvlc,
		gopacket.Payload(npdu),
	); err != nil {
		return -1, err
	}

	return c.PacketConn.WriteTo(bf.Bytes(), dst.UDPAddr())
}

func (c *BIPConn) ReadFrom(npdu []byte) (n int, addr net.Addr, err error) {
	select {
	case <-c.done:
		return -1, nil, &net.OpError{Op: "read", Net: addr.Network(), Source: c.laddr, Addr: addr, Err: syscall.EPIPE}
	case err := <-c.errCh:
		return -1, nil, err
	case fr := <-c.recvCh:
		copy(npdu, fr.payload)
		return len(fr.payload), fr.addr, nil
	}
}

func (c *BIPConn) Close() error {
	close(c.done)
	err := c.PacketConn.Close()
	close(c.recvCh)
	close(c.errCh)
	return err
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
