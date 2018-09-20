package layers

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mailru/easyjson/buffer"
	"github.com/stitchcula/bacnet-go"
)

type NPDUPriority byte

const (
	NPDUPriorityNormal NPDUPriority = iota
	NPDUPriorityUrgent
	NPDUPriorityCriticalEquipment
	NPDUPriorityLifeSafety
)

type NPDUFlag byte

const (
	NPDUMaskAPDU           NPDUFlag = 1 << 7
	NPDUMaskDst            NPDUFlag = 1 << 5
	NPDUMaskSrc            NPDUFlag = 1 << 3
	NPDUMaskExpectingReply NPDUFlag = 1 << 2
)

func (f NPDUFlag) HasAPDU() bool {
	return f&NPDUMaskAPDU > 0
}

func (f NPDUFlag) HasDst() bool {
	return f&NPDUMaskDst > 0
}

func (f NPDUFlag) HasSrc() bool {
	return f&NPDUMaskSrc > 0
}

func (f NPDUFlag) ExpectingReply() bool {
	return f&NPDUMaskExpectingReply > 0
}

func (f NPDUFlag) Priority() NPDUPriority {
	return NPDUPriority(f & 3)
}

type NPDU struct {
	layers.BaseLayer

	ProtocolVersion byte
	Flags           NPDUFlag
	Dst, Src        *bacnet.Addr
	HopCount        byte
	MessageType     byte
	VendorID        bacnet.VendorID
}

func (npdu *NPDU) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {

}

func (npdu *NPDU) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	npdu.ProtocolVersion = data[0]
	npdu.Flags = NPDUFlag(data[1])

	sk := 2
	if npdu.Flags.HasDst() {
		npdu.Dst = &bacnet.Addr{}
		sk += npdu.Dst.DecodeFromBytes(data[sk:])
	}

	if npdu.Flags.HasSrc() {
		npdu.Src = &bacnet.Addr{}
		sk += npdu.Src.DecodeFromBytes(data[sk:])
	}

	if npdu.Flags.HasDst() {
		npdu.HopCount = data[sk]
		sk++
	} else {
		npdu.HopCount = 0
	}

	if npdu.Flags.HasAPDU() {
		npdu.MessageType = data[sk]
		sk++
		if npdu.MessageType > 0x80 {
			npdu.VendorID = bacnet.VendorID(binary.BigEndian.Uint16(data[sk : sk+2]))
			sk = sk + 2
		}
	}

	npdu.BaseLayer.Contents = data[:sk]
	npdu.BaseLayer.Payload = data[sk:]

	return nil
}

func (npdu *NPDU) NetworkFlow() gopacket.Flow {
	return gopacket.NewFlow(EndpointBACnetNPDU, npdu.Src.Bytes(), npdu.Dst.Bytes())
}

// LayerType returns LayerTypeBACnetVLC
func (npdu *NPDU) LayerType() gopacket.LayerType { return LayerTypeBACnetNPDU }

func decodeBACnetNPDU(data []byte, p gopacket.PacketBuilder) error {
	npdu := &NPDU{}
	err := npdu.DecodeFromBytes(data, p)
	p.AddLayer(npdu)
	p.SetNetworkLayer(npdu)
	if err != nil {
		return err
	}

	return p.NextDecoder(LayerTypeBACnetAPDU)
}
