package layers

import "github.com/google/gopacket"

const ()

var (
	// LayerTypeBACnetLPDU = gopacket.RegisterLayerType(1233, gopacket.LayerTypeMetadata{Name: "BACnetLPDU", Decoder: gopacket.DecodeFunc(decodeBACnetLPDU)})
	LayerTypeBACnetNPDU = gopacket.RegisterLayerType(1234, gopacket.LayerTypeMetadata{Name: "BACnetNPDU", Decoder: gopacket.DecodeFunc(decodeBACnetNPDU)})
	LayerTypeBACnetAPDU = gopacket.RegisterLayerType(1235, gopacket.LayerTypeMetadata{Name: "BACnetAPDU", Decoder: gopacket.DecodeFunc(decodeBACnetAPDU)})
	LayerTypeBACnetVLC  = gopacket.RegisterLayerType(1236, gopacket.LayerTypeMetadata{Name: "BACnetVLC", Decoder: gopacket.DecodeFunc(decodeBACnetVLC)})
)

var (
	EndpointBACnetNPDU = gopacket.RegisterEndpointType(1234, gopacket.EndpointTypeMetadata{Name: "BACnetNPDU", Formatter: func([]byte) string {
		return "BACnetNPDU"
	}})
	EndpointBACnetVLC = gopacket.RegisterEndpointType(1236, gopacket.EndpointTypeMetadata{Name: "BACnetVLC", Formatter: func([]byte) string {
		return "BACnetVLC"
	}})
)

// setByteMask takes an input in, and make a bit either 0, or 1 depending on the
// input boolean and mask
func SetByteMask(in byte, b bool, mask byte) byte {
	if b {
		return in | mask
	} else {
		var m byte = 0xFF
		m = m - mask
		return in & m
	}
}
