package handler

import (
	"bytes"
	"fmt"

	"github.com/LiteAdminProd/BedrockProxy/src/logger"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
	msg := fmt.Sprintf("Player connected: %s | ip: %s | xuid: %s | uuid: %s | device: %s", nick, addr, xuid, uuid, device)
	logger.Info(msg)
}

// TODO: make it works
func Text(payload []byte) {
	reader := protocol.NewReader(bytes.NewBuffer(payload), 0)
	var packet packet.Text
	packet.Marshal(reader)
	var msg string
	switch packet.TextType {
	case 1:
		msg = "<" + packet.SourceName + "> " + packet.Message
	case 7:
		msg = "<" + packet.SourceName + "> " + packet.Message
	}
	logger.Info(msg)
}

func Disconnect(conn *minecraft.Conn) {
	logger.Info("Player disconnected: " + conn.IdentityData().DisplayName)
}

func Transfer(payload []byte) {
	reader := protocol.NewReader(bytes.NewBuffer(payload), 0)
	var packet packet.Transfer
	packet.Marshal(reader)
	logger.Info("Player transfered to", packet.Address, packet.Port)
}
