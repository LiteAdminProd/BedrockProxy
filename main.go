package main

import (
	"fmt"
	"log"
	"net"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {

	const addr = "0.0.0.0:228"

	status, err := minecraft.NewForeignStatusProvider(addr)
	if err != nil {
		log.Panic(err)
	}
	listen := minecraft.ListenConfig{
		StatusProvider:         status,
		AuthenticationDisabled: false,
		MaximumPlayers:         10,
		PacketFunc:             handle,
	}
	listener, err := listen.Listen("raknet", addr)
	if err != nil {
		log.Panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("conn: %v\n", conn.RemoteAddr())
	}

}

func handle(header packet.Header, payload []byte, src net.Addr, dst net.Addr) {
	log.Print(header.PacketID)
}
