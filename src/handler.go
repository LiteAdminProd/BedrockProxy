package handler

import (
	"github.com/sandertv/gophertunnel/minecraft"
	//"github.com/sandertv/gophertunnel/minecraft/protocol"
	//"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"log"
)

func LoginMessage(conn *minecraft.Conn) {
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
	addr := conn.RemoteAddr().String()
	log.Printf("Player login: %s | ip: %s | xuid: %s | uuid: %s | device: %s", addr, nick, xuid, uuid, device)
}

// TODO: make it works
func ChatMessage(payload *[]byte) {
	// var io protocol.IO
	// io.Bytes(payload)

	// var text packet.Text
	// text.Marshal(io)

	// log.Print("Type: ", (*payload)[0])
	// log.Print(text.Message)
}
