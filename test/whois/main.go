package main

import (
	"fmt"
	"github.com/stitchcula/bacnet-go/datalink"
)

func main() {
	fmt.Println(datalink.NewBVLCConn("192.168.50.148"))
}
