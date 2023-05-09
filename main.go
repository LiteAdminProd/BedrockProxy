package main

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const LocalAddress = "localhost:20777"
const SendToAddress = "localhost:19132"

func main() {
	status, err := minecraft.NewForeignStatusProvider(SendToAddress)
	if err != nil {
		log.Panic(err)
	}
	listen := minecraft.ListenConfig{
		StatusProvider:         status,
		AuthenticationDisabled: false,
		MaximumPlayers:         10,
		PacketFunc:             handle,
	}
	listener, err := listen.Listen("raknet", LocalAddress)
	if err != nil {
		log.Panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn.(*minecraft.Conn), listener)
	}

}

func handle(header packet.Header, payload []byte, src net.Addr, dst net.Addr) {
	log.Print(src, " -> ", dst, "|", header.PacketID)
}

func handleConn(conn *minecraft.Conn, listener *minecraft.Listener) {
	serverConn, err := minecraft.Dialer{
		ClientData:   conn.ClientData(),
		IdentityData: conn.IdentityData(),
	}.Dial("raknet", SendToAddress)
	if err != nil {
		panic(err)
	}
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	go func() {
		defer listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
		}
	}()
	go func() {
		defer serverConn.Close()
		defer listener.Disconnect(conn, "connection lost")
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}
