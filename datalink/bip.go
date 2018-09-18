package datalink

// #cgo CFLAGS: -I ../bacnet-stack/include
// #cgo windows CFLAGS: -I ../bacnet-stack/ports/win32
// #include "datalink.h"
// #cgo LDFLAGS: -L ../bacnet-stack/lib -lbacnet
// #cgo windows LDFLAGS: -lws2_32 -lwsock32 -liphlpapi
import "C"
import (
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"
)

type BVLCConn struct {
}

func NewBVLCConn(ifn string) (c *BVLCConn, err error) {
	CIfm := C.CString(ifn)
	defer C.free(unsafe.Pointer(CIfm))
	if !bool(C.bip_init(CIfm)) {
		return nil, errors.New("bip_init failed")
	}

	broadcast := &C.struct_BACnet_Device_Address{}
	C.datalink_get_broadcast_address(broadcast)
	fmt.Println(broadcast)

	return &BVLCConn{}, nil
}

func (c *BVLCConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return -1, nil, nil
}

func (c *BVLCConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return -1, nil
}

func (c *BVLCConn) Close() error {
	return nil
}

func (c *BVLCConn) LocalAddr() net.Addr {
	return nil
}

func (c *BVLCConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *BVLCConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *BVLCConn) SetWriteDeadline(t time.Time) error {
	return nil
}
