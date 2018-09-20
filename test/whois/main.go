package main

import (
	"fmt"
	"net"
)

func main() {
	uni, _ := net.InterfaceAddrs()
	for i := range uni {
		ip, _, err := net.ParseCIDR(uni[i].String())
		if err != nil || ip.IsUnspecified() || ip.IsLoopback() || ip.To4() == nil {
			continue
		}
		fmt.Println(ip)
	}
}
