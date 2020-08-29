package main

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type Hikari struct {
	parent *Utsutsu
	nickName string

	pcGame   net.PacketConn
	pcYume   net.PacketConn
	GameAddr net.Addr
	YumeAddr net.Addr
}

func NewHikari(parent *Utsutsu) *Hikari {
	h := new(Hikari)
	h.parent = parent
	h.nickName = "unknown"
	h.GameAddr, _ = net.ResolveUDPAddr("udp4", "127.0.0.1:10800")

	return h
}

func (h *Hikari) Start() {

	yumeAddr := h.parent.Tunnel.zNewLink()
	h.YumeAddr, _ = net.ResolveUDPAddr("udp4", yumeAddr)

	pcGame, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		h.parent.logger.WithFields(logrus.Fields{
			"scope": "hikari/Start",
		}).Fatal(err)
	}

	h.pcGame = pcGame

	pcYume, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		h.parent.logger.WithFields(logrus.Fields{
			"scope": "hikari/Start",
		}).Fatal(err)
	}

	h.pcYume = pcYume

	go h.keepAlive()
	go h.handleGameConnection()
	go h.handleYumeConnection()

}

func (h *Hikari) Stop() {
	h.pcGame.Close()
	h.pcYume.Close()

	h.parent.Tunnel.zDeleteLink()
}

func (h *Hikari) keepAlive() {
	for {
		_, err := h.pcYume.WriteTo([]byte("PHANTOM"), h.YumeAddr)

		if err != nil {
			//h.parent.logger.WithFields(logrus.Fields{
			//	"scope": "hikari/keepAlive",
			//}).Warn(err)
			break
		}

		time.Sleep(10 * time.Second)

	}
}

func (h *Hikari) handleGameConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := h.pcGame.ReadFrom(buf)
		fmt.Println("G: ", buf[:3])

		if err != nil {
			//h.parent.logger.WithFields(logrus.Fields{
			//	"scope": "hikari/handleGameConnection",
			//}).Warn(err)
			break
		}
		h.pcYume.WriteTo(buf[:n], h.YumeAddr)
	}
}

func (h *Hikari) handleYumeConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := h.pcYume.ReadFrom(buf)
		fmt.Println(buf[:3])

		if err != nil {
			//h.parent.logger.WithFields(logrus.Fields{
			//	"scope": "hikari/handleYumeConnection",
			//}).Warn(err)
			break
		}
		h.pcGame.WriteTo(buf[:n], h.GameAddr)
	}
}
