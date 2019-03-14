package main

import (
	"fmt"
	"github.com/stitchcula/bacnet-go"
	"github.com/stitchcula/bacnet-go/datalink"
	"github.com/stitchcula/bacnet-go/layers"
	"github.com/stitchcula/bacnet-go/network"
)

func main() {
	conn, err := network.ListenPacket(datalink.BIP, "192.168.30.212:47808")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.WriteTo([]byte{0x10, 0x08}, bacnet.BroadcastAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		data := make([]byte, 0, layers.MaxAPDU)
		_, addr, err := conn.ReadFrom(data)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s %x\r\n", addr.String(), data)
	}
}
