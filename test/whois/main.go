package main

import (
	"fmt"
	"github.com/stitchcula/bacnet-go/datalink"
	"net"
)

func main() {
	link, err := datalink.NewBIPConn("192.168.50.148")
	if err != nil {
		fmt.Println(err)
		return
	}
	laddr := link.LocalAddr().(*net.UDPAddr)
	fmt.Println(laddr)
	fmt.Println(laddr.IP.IsUnspecified())

}
