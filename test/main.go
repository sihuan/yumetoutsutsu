package main

import (
	"bytes"
	"fmt"
	"net"
)

func main() {
	pcGuest, _ := net.ListenPacket("udp4", "0.0.0.0:0")
	fmt.Println(pcGuest.LocalAddr().String())
	pcHost, _ := net.ListenPacket("udp4", "127.0.0.1:10800")
	fmt.Println(pcHost.LocalAddr().String())

	ch := make(chan int)

	go func() {
		YamiAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:31106")
		if err != nil {
			fmt.Println(err)
			return
		}

		pcGuest.WriteTo([]byte{1,2,3}, YamiAddr)
		fmt.Println("Guest Send To Yami")

		buf := make([]byte, 1536)
		n, _, _ := pcGuest.ReadFrom(buf)
		fmt.Println("Guest Read From Yami", buf[:n])

		if bytes.Equal(buf[:n], []byte{1,2,3}) {
			ch <- 1
		}


	}()

	go func() {
		buf := make([]byte, 1536)

		n, addr, _ := pcHost.ReadFrom(buf)
		fmt.Println("Host Read From Hikari", buf[:n])

		pcHost.WriteTo(buf[:n], addr)
		fmt.Println("Host Send To Hikari")
	}()

	<- ch
	return
}
