package datalink

import (
	"net"
	"strconv"
)

const (
	DefaultPort = 0xBAC0
)

type BIPConn struct {
	net.PacketConn
	laddr      *net.UDPAddr
	broadcasts []net.IP
}

func NewBIPConn(ifn string) (c *BIPConn, err error) {
	if net.ParseIP(ifn) != nil {
		ifn = ifn + ":" + strconv.Itoa(DefaultPort)
	} else if _, _, err = net.SplitHostPort(ifn); err != nil {
		ifn = ":" + strconv.Itoa(DefaultPort)
	}

	c = &BIPConn{}
	c.PacketConn, err = net.ListenPacket("udp", ifn)
	if err != nil {
		return nil, err
	}

	c.laddr = c.PacketConn.LocalAddr().(*net.UDPAddr)
	_, err = c.ResolveBroadcasts(c.laddr)
	if err != nil {
		return nil, err
	}

	return
}

func (c *BIPConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if addr == nil {
		return c.Broadcast(p)
	}
	return c.PacketConn.WriteTo(p, addr)
}

// Broadcast write p to all c.broadcasts addresses
// TODO: should use default port of laddr?
func (c *BIPConn) Broadcast(p []byte) (n int, err error) {
	for i := range c.broadcasts {
		n, err = c.PacketConn.WriteTo(p, &net.UDPAddr{
			IP:   c.broadcasts[i],
			Port: c.laddr.Port,
		})
		if err != nil {
			break
		}
	}

	return
}

// ResolveBroadcasts resolve the broadcast IPs to c.broadcasts
func (c *BIPConn) ResolveBroadcasts(laddr *net.UDPAddr) (addrs []net.IP, err error) {
	uni, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for i := range uni {
		ip, msk, err := net.ParseCIDR(uni[i].String())
		// TODO: ipv4 only?
		if err != nil || ip.To4() == nil || (!laddr.IP.IsUnspecified() && !laddr.IP.Equal(ip)) {
			continue
		}
		broadcast := net.IP(make([]byte, 4))
		for i := range broadcast {
			broadcast[i] = msk.IP[i] | ^msk.Mask[i]
		}
		c.broadcasts = append(c.broadcasts, broadcast)
	}

	return c.broadcasts, nil
}
