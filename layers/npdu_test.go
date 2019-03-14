package layers

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/stitchcula/bacnet-go"
	"testing"
)

func TestNPDU_SerializeTo(t *testing.T) {
	npdu := &NPDU{
		ProtocolVersion:       NPDUProtocolVersion,
		Dst:                   bacnet.BroadcastAddr,
		Src:                   nil,
		IsNetworkLayerMessage: false,
		ExpectingReply:        false,
		Priority:              NPDUPriorityNormal,
		HopCount:              NPDUDefaultHopCount,
	}

	bf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(bf, gopacket.SerializeOptions{},
		npdu,
		gopacket.Payload([]byte{0x10, 0x08}),
	)

	fmt.Printf("%x\r\n", bf.Bytes())
}
