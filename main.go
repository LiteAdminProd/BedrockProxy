package main

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const LocalAddress = "0.0.0.0:20777"
const SendToAddress = "0.0.0.0:228"
const debug = true

func main() {
	status, err := minecraft.NewForeignStatusProvider(SendToAddress)
	if err != nil {
		log.Panic(err)
	}
	listen := minecraft.ListenConfig{
		StatusProvider:         status,
		AuthenticationDisabled: false,
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
	if debug {
		log.Print(src, " -> ", dst, "|", header.PacketID)
	}
	
}

func handleConn(conn *minecraft.Conn, listener *minecraft.Listener) {
	serverConn, err := minecraft.Dialer{
		ClientData:   conn.ClientData(),
		IdentityData: conn.IdentityData(),
	}.Dial("raknet", SendToAddress)
	if err != nil {
		panic(err)
	}

	var device string
	const WinTitleID = "896928775"
	const PhoneTitleID = "1739947436"
	if conn.IdentityData().TitleID == WinTitleID {
		device = "WIN"
	} else if conn.IdentityData().TitleID == PhoneTitleID {
		device = "PHONE"
	} else {
		device = "OTHER"
	}
	nick := conn.IdentityData().DisplayName
	xuid := conn.IdentityData().XUID
	uuid := conn.IdentityData().Identity
	log.Printf("Player login: %s | xuid: %s | uuid: %s | device: %s", nick, xuid, uuid, device)

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
