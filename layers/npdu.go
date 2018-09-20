package layers

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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

func (f *NPDUFlag) setDst(b bool) {
	*f = NPDUFlag(SetByteMask(byte(*f), b, byte(NPDUMaskDst)))
}

func (f NPDUFlag) HasDst() bool {
	return f&NPDUMaskDst > 0
}

func (f *NPDUFlag) setSrc(b bool) {
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
	Flags           NPDUFlag
	Dst, Src        *bacnet.Addr
	HopCount        byte
	MessageType     byte
	VendorID        bacnet.VendorID
}

func (npdu *NPDU) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(2)
	if err != nil {
		return err
	}
	bytes[0] = npdu.ProtocolVersion

	if npdu.Dst != nil || npdu.Dst.Net != 0 {
		npdu.Flags.setDst(true)

		raddr := npdu.Dst.Bytes()
		dst, err := b.PrependBytes(len(raddr))
		if err != nil {
			return err
		}
		copy(dst, raddr)
	}

	if npdu.Src != nil || npdu.Src.Net != 0 {
		npdu.Flags.setSrc(true)

		laddr := npdu.Src.Bytes()
		src, err := b.PrependBytes(len(laddr))
		if err != nil {
			return err
		}
		copy(src, laddr)
	}

	bytes[1] = byte(npdu.Flags)

	if npdu.Flags.HasDst() {
		bytes, err = b.PrependBytes(1)
		if err != nil {
			return err
		}
		bytes[0] = npdu.HopCount
	}

	if npdu.Flags.IsNetworkLayerMessage() {
		bytes, err = b.PrependBytes(1)
		if err != nil {
			return err
		}
		bytes[0] = npdu.MessageType
		if npdu.MessageType > 0x80 {
			bytes, err = b.PrependBytes(2)
			if err != nil {
				return err
			}
			binary.BigEndian.PutUint16(bytes, uint16(npdu.VendorID))
		}
	}

	return nil
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

	if npdu.Flags.IsNetworkLayerMessage() {
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
