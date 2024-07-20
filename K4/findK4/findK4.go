package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

	fmt.Println("findK4 V1.0 - use UDP broadcast to get all active K4s radios to say their ID")

	const K4finder string = "findk4"
	// var recvAddr net.Addr

	pc, err := net.ListenPacket("udp4", ":51")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	addr, err := net.ResolveUDPAddr("udp4", "192.168.21.255:9100")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("bcast addr: %+v finder string: %s\n", addr, K4finder)

	_, err = pc.WriteTo([]byte(K4finder), addr)

	if err != nil {
		panic(err)
	}

	pc.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 1024)
	for {
		n, recvAddr, err := pc.ReadFrom(buf)

		if err != nil {
			panic(err)
		}
		fmt.Printf("rcved K4 finder:%s %s\n", recvAddr, buf[:n])

	}

}
