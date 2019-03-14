package layers

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stitchcula/bacnet-go"
)

const NPDUProtocolVersion uint8 = 1

const NPDUDefaultHopCount uint8 = 255

type NPDUPriority byte

const (
	NPDUPriorityNormal NPDUPriority = iota
	NPDUPriorityUrgent
	NPDUPriorityCriticalEquipment
	NPDUPriorityLifeSafety
)

type NPDUFlag byte

const (
	NPDUMaskNetworkLayerMessage NPDUFlag = 1 << 7
	NPDUMaskDst                 NPDUFlag = 1 << 5
	NPDUMaskSrc                 NPDUFlag = 1 << 3
	NPDUMaskExpectingReply      NPDUFlag = 1 << 2
)

func (f *NPDUFlag) SetNetworkLayerMessage(b bool) {
	*f = NPDUFlag(SetByteMask(byte(*f), b, byte(NPDUMaskNetworkLayerMessage)))
}

func (f NPDUFlag) IsNetworkLayerMessage() bool {
	return f&NPDUMaskNetworkLayerMessage > 0
}

func (f *NPDUFlag) SetDst(b bool) {
	*f = NPDUFlag(SetByteMask(byte(*f), b, byte(NPDUMaskDst)))
}

func (f NPDUFlag) HasDst() bool {
	return f&NPDUMaskDst > 0
}

func (f *NPDUFlag) SetSrc(b bool) {
	*f = NPDUFlag(SetByteMask(byte(*f), b, byte(NPDUMaskSrc)))
}

func (f NPDUFlag) HasSrc() bool {
	return f&NPDUMaskSrc > 0
}

func (f *NPDUFlag) SetExpectingReply(b bool) {
	*f = NPDUFlag(SetByteMask(byte(*f), b, byte(NPDUMaskExpectingReply)))
}

func (f NPDUFlag) ExpectingReply() bool {
	return f&NPDUMaskExpectingReply > 0
}

func (f *NPDUFlag) SetPriority(p NPDUPriority) {
	*f |= NPDUFlag(p)
}

func (f NPDUFlag) Priority() NPDUPriority {
	return NPDUPriority(f & 3)
}

type NPDU struct {
	layers.BaseLayer
	ProtocolVersion byte

	IsNetworkLayerMessage bool
	ExpectingReply        bool
	Priority              NPDUPriority
	flags                 NPDUFlag

	Dst, Src    *bacnet.Addr
	HopCount    byte
	MessageType byte
	VendorID    bacnet.VendorID
}

func (npdu *NPDU) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	byt := make([]byte, 2, MaxNPDU)
	byt[0] = npdu.ProtocolVersion

	npdu.flags.SetNetworkLayerMessage(npdu.IsNetworkLayerMessage)
	npdu.flags.SetExpectingReply(npdu.ExpectingReply)
	npdu.flags.SetPriority(npdu.Priority)

	if npdu.Dst != nil && npdu.Dst.Net != 0 {
		npdu.flags.SetDst(true)

		raddr := npdu.Dst.Bytes()
		byt = append(byt, raddr...)
	}

	if npdu.Src != nil && npdu.Src.Net != 0 {
		npdu.flags.SetSrc(true)

		laddr := npdu.Src.Bytes()
		byt = append(byt, laddr...)
	}

	byt[1] = byte(npdu.flags)

	if npdu.flags.HasDst() {
		byt = append(byt, npdu.HopCount)
	}

	if npdu.flags.IsNetworkLayerMessage() {
		byt = append(byt, npdu.MessageType)
		if npdu.MessageType > 0x80 {
			byt = append(byt, byte(npdu.VendorID>>8), byte(npdu.VendorID))
		}
	}

	dst, err := b.PrependBytes(len(byt))
	if err != nil {
		return err
	}
	copy(dst, byt)

	return nil
}

func (npdu *NPDU) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	npdu.ProtocolVersion = data[0]
	npdu.flags = NPDUFlag(data[1])

	npdu.IsNetworkLayerMessage = npdu.flags.IsNetworkLayerMessage()
	npdu.ExpectingReply = npdu.flags.ExpectingReply()
	npdu.Priority = npdu.flags.Priority()

	sk := 2
	if npdu.flags.HasDst() {
		npdu.Dst = &bacnet.Addr{}
		sk += npdu.Dst.DecodeFromBytes(data[sk:])
	}

	if npdu.flags.HasSrc() {
		npdu.Src = &bacnet.Addr{}
		sk += npdu.Src.DecodeFromBytes(data[sk:])
	}

	if npdu.flags.HasDst() {
		npdu.HopCount = data[sk]
		sk++
	} else {
		npdu.HopCount = 0
	}

	if npdu.flags.IsNetworkLayerMessage() {
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

	// TODO: ???
	return p.NextDecoder(gopacket.LayerTypePayload)
}
