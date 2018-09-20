package layers

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type BVLCFunction byte

const (
	BVLCResult BVLCFunction = iota
	BVLCWriteBroadcastDistTable
	BVLCReadBroadcastDistTable
	BVLCReadBroadcastDistTableAck
	BVLCForwardedNPDU
	BVLCRegisterForeignDevice
	BVLCReadForeignDeviceTable
	BVLCReadForeignDeviceTableAck
	BVLCDeleteForeignDeviceTableEntry
	BVLCDistributeBroadcastToNetwork
	BVLCOriginalUnicastNPDU
	BVLCOriginalBroadcastNPDU
	BVLCSecureBVLL
	MaxBVLCFunction
)

// BVLCTypeBIP is the only valid type for the BVLC layer as of 2002.
// Additional types may be added in the future
const BVLCTypeBIP = 0x81

// BVLC is the layer for BACnet Virtual Link Control
type BVLC struct {
	layers.BaseLayer
	Type     byte
	Function BVLCFunction
	length   uint16
}

func (vlc *BVLC) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(4)
	if err != nil {
		return err
	}
	vlc.length = uint16(4 + len(b.Bytes()))

	bytes[0] = BVLCTypeBIP // vlc.Type
	bytes[1] = byte(vlc.Function)
	binary.BigEndian.PutUint16(bytes[2:], vlc.length)

	return nil
}

func (vlc *BVLC) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	vlc.Type = data[0]
	if vlc.Type != BVLCTypeBIP {
		return fmt.Errorf("invalid BVLC.Type %d of %d", BVLCTypeBIP, vlc.Type)
	}
	vlc.Function = BVLCFunction(data[1])
	vlc.length = binary.BigEndian.Uint16(data[2:4])

	if len(data) != int(vlc.length) {
		return fmt.Errorf("invalid BVLC.length %d of %d", vlc.length, len(data))
	}
	vlc.Contents = data[:4]
	vlc.Payload = data[4:]
	return nil
}

// LayerType returns LayerTypeBACnetVLC
func (vlc *BVLC) LayerType() gopacket.LayerType { return LayerTypeBACnetVLC }

func (vlc *BVLC) LinkFlow() gopacket.Flow {
	return gopacket.NewFlow(EndpointBACnetVLC, nil, nil)
}

func decodeBACnetVLC(data []byte, p gopacket.PacketBuilder) error {
	vlc := &BVLC{}
	err := vlc.DecodeFromBytes(data, p)
	p.AddLayer(vlc)
	p.SetLinkLayer(vlc)
	if err != nil {
		return err
	}
	return p.NextDecoder(LayerTypeBACnetNPDU)
}
