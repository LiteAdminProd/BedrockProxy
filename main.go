package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"sync"

	handler "github.com/LiteAdminProd/BedrockProxy/src"
	"github.com/LiteAdminProd/BedrockProxy/src/logger"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var conf Config

func main() {
	logger.Info("Loading BedrockProxy")
	logger.Info("Loading config.json")
	conf = LoadConfiguration()
	logger.Info("Requesting motd from " + conf.SendToAddress)
	status, err := minecraft.NewForeignStatusProvider(conf.SendToAddress)
	if err != nil {
		log.Panic(err)
	}
	listen := minecraft.ListenConfig{
		StatusProvider:         status,
		AuthenticationDisabled: false,
		PacketFunc:             handle,
	}
	logger.Info("BedrockProxy listening on " + conf.LocalAddress)
	listener, err := listen.Listen("raknet", conf.LocalAddress)
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
	if conf.Debug {
		logger.Info(src, " -> ", dst, "|", header.PacketID)
	}
	switch header.PacketID {
	// 0x09 is a chat message packet
	case 9:
		addr, err := net.ResolveUDPAddr("udp", conf.LocalAddress)
		if err != nil {
			log.Panic(err)
		}

		if addr.Port == src.(*net.UDPAddr).Port {
			handler.Text(payload)
		}
	case 85:
		addr, err := net.ResolveUDPAddr("udp", conf.LocalAddress)
		if err != nil {
			log.Panic(err)
		}

		if addr.Port == src.(*net.UDPAddr).Port {
			handler.Transfer(payload)
		}
	}
}

func handleConn(conn *minecraft.Conn, listener *minecraft.Listener) {
	serverConn, err := minecraft.Dialer{
		ClientData:   conn.ClientData(),
		IdentityData: conn.IdentityData(),
	}.Dial("raknet", conf.SendToAddress)
	if err != nil {
		logger.Error(err.Error())
	}
	handler.LoginMessage(conn)

	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			logger.Error(err.Error())
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			logger.Error(err)
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
				handler.Disconnect(conn)
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

type Config struct {
	LocalAddress  string `json:"localAddress"`
	SendToAddress string `json:"sendToAddress"`
	Debug         bool   `json:"debug"`
}

func LoadConfiguration() Config {
	var config Config
	file := "config.json"

	configFile, err := os.Open(file)
	if os.IsNotExist(err) {
		logger.Warn("config.json not found, creating new one")
		file, err := os.Create("config.json")
		if err != nil {
			log.Fatal(err)
		}

		defaultConf := Config{
			LocalAddress:  "0.0.0.0:19132",
			SendToAddress: "0.0.0.0:19133",
			Debug:         false,
		}
		marshalDefaultConf, _ := json.MarshalIndent(defaultConf, "", "    ")
		file.Write(marshalDefaultConf)
		file.Close()
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	configFile.Close()
	return config
}
