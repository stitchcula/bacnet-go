package datalink

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go/layers"
	"net"
	"strconv"
	"time"
)

const (
	DefaultPort = 0xBAC0
)

type singleBIPConn struct {
	net.PacketConn
	msk       *net.IPNet
	laddr     *net.UDPAddr
	broadcast net.IP
}

func newSingleBIPConn(ifn string) (c *singleBIPConn, err error) {
	c = &singleBIPConn{}
	c.PacketConn, err = net.ListenPacket("udp", ifn)
	if err != nil {
		return nil, err
	}

	c.laddr = c.PacketConn.LocalAddr().(*net.UDPAddr)
	if err = c.resolveBroadcast(c.laddr); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *singleBIPConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if addr == nil {
		return c.Broadcast(p)
	}

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{},
		&layers.BVLC{},
		&layers.NPDU{},
		gopacket.Payload(p),
	)

	return c.PacketConn.WriteTo(buf.Bytes(), addr)
}

func (c *singleBIPConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return c.PacketConn.ReadFrom(p)
}

// resolveBroadcast resolve the broadcast IP to c.broadcast
func (c *singleBIPConn) resolveBroadcast(laddr *net.UDPAddr) error {
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
		c.broadcast = net.IP(make([]byte, 4))
		for i := range c.broadcast {
			c.broadcast[i] = msk.IP[i] | ^msk.Mask[i]
		}
	}

	if c.broadcast == nil {
		return fmt.Errorf("can not bind network interface")
	}

	return nil
}

type BIPConn struct {
	laddr net.Addr
	conns []*singleBIPConn
}

func NewBIPConn(ifn string) (*BIPConn, error) {
	h, port, err := net.SplitHostPort(ifn)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(h)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 %s", h)
	}

	laddr, _ := net.ResolveUDPAddr("udp", ifn)
	c := &BIPConn{laddr: laddr}

	if !ip.IsUnspecified() {
		sc, err := newSingleBIPConn(ifn)
		if err != nil {
			return nil, err
		}
		c.conns = append(c.conns, sc)
		return c, nil
	}

	uni, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for i := range uni {
		ip, _, err := net.ParseCIDR(uni[i].String())
		if err != nil || ip.IsUnspecified() || ip.IsLoopback() || ip.To4() == nil {
			continue
		}

		sc, err := newSingleBIPConn(net.JoinHostPort(ip.String(), port))
		if err != nil {
			return nil, err
		}
		c.conns = append(c.conns, sc)
	}

	return c, nil
}

func (c *BIPConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return
}

func (c *BIPConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	// TODO: read from channel
	return
}

func (c *BIPConn) Close() (err error) {
	for i := range c.conns {
		if er := c.conns[i].Close(); er != nil {
			err = er
		}
	}
	return err
}

func (c *BIPConn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *BIPConn) SetDeadline(t time.Time) (err error) {
	for i := range c.conns {
		if er := c.conns[i].SetDeadline(t); er != nil {
			err = er
		}
	}
	return err
}

func (c *BIPConn) SetReadDeadline(t time.Time) (err error) {
	for i := range c.conns {
		if er := c.conns[i].SetReadDeadline(t); er != nil {
			err = er
		}
	}
	return err
}

func (c *BIPConn) SetWriteDeadline(t time.Time) (err error) {
	for i := range c.conns {
		if er := c.conns[i].SetWriteDeadline(t); er != nil {
			err = er
		}
	}
	return err
}
