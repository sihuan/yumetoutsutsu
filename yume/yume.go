package main

import (
	"encoding/gob"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Yume struct {
	PublicAddr string

	logger   *logrus.Logger
	Tunnel   *Tunnel
	utsutsus []UtsutsuInfo
	links    []Link
}

func NewYume() *Yume {

	y := new(Yume)

	//y.PublicAddr = "39.106.101.255"
	y.PublicAddr = "127.0.0.1"

	y.logger = logrus.New()

	y.Tunnel = NewTunnel(y)

	y.utsutsus = make([]UtsutsuInfo, 0)
	y.links = make([]Link, 0)

	y.Tunnel.OnRequestReceived.SubscribeAsync("utsutsu.init", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		var nickName string
		decoder.Decode(&nickName)

		ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

		utsutsu := NewUtsutsuInfo(conn.RemoteAddr().String(), ip, conn, session)

		y.utsutsus = append(y.utsutsus, utsutsu)

		encoder.Encode(utsutsu.Addr)

	}, false)

	y.Tunnel.OnRequestReceived.SubscribeAsync("links", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {
		// fmt.Println(y.links)
		y.recycle()
		encoder.Encode(y.links)

	}, false)

	y.Tunnel.OnRequestReceived.SubscribeAsync("nlink", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		var utsutsuAddr string
		decoder.Decode(&utsutsuAddr)
		var nickName string
		decoder.Decode(&nickName)

		l := y.NewLink(utsutsuAddr, nickName)
		y.links = append(y.links, *l)

		encoder.Encode(l.HikariAddr())

	}, false)

	y.Tunnel.OnRequestReceived.SubscribeAsync("dlink", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		var utsutsuAddr string
		decoder.Decode(&utsutsuAddr)

		idx := y.findLinkAndRemove(utsutsuAddr)
		encoder.Encode(idx)

	}, false)

	y.Tunnel.OnPeerConnected.SubscribeAsync("", func(conn *kcp.UDPSession, session *yamux.Session) {

		y.logger.WithFields(logrus.Fields{
			"scope": "yume/handlePeerConnected",
		}).Info("handle peer connected")

	}, false)

	y.Tunnel.OnPeerDisconnected.SubscribeAsync("", func(conn *kcp.UDPSession, session *yamux.Session) {

		y.logger.WithFields(logrus.Fields{
			"scope": "yume/handlePeerDisconnected",
		}).Info("handle peer disconnected")

		y.findUtsutsuAndRemove(conn, session)

	}, false)

	return y

}

func (y *Yume) findUtsutsuAndRemove(conn *kcp.UDPSession, session *yamux.Session) int {

	idx := -1

	for i, utsutsu := range y.utsutsus {
		if utsutsu.conn == conn && utsutsu.session == session {
			idx = i
			break
		}
	}

	if idx >= 0 {
		y.findLinkAndRemove(y.utsutsus[idx].Addr)
		y.utsutsus = append(y.utsutsus[0:idx], y.utsutsus[idx+1:]...)
	}

	return idx

}

func (y *Yume) findLinkAndRemove(utsutsuAddr string) int {

	idx := -1

	for i, link := range y.links {
		if link.utsutsuAddr == utsutsuAddr {
			idx = i
			break
		}
	}

	if idx >= 0 {
		y.links[idx].Stop()
		y.links = append(y.links[0:idx], y.links[idx+1:]...)
	}

	return idx

}

func (y *Yume) recycle() {

		for i := len(y.links) - 1; i >= 0; i-- {
			link := y.links[i]

			if link.HikariExpired() {
				y.logger.WithFields(logrus.Fields{
					"scope":          "yume/recycle",
					"linkHikariAddr": link.HikariAddr(),
				}).Info("recycle link, hikari dead")
				link.Stop()
				y.links = append(y.links[:i], y.links[i+1:]...)

			}

			if link.YamiExpired() {
				y.logger.WithFields(logrus.Fields{
					"scope":        "yume/recycle",
					"linkYamiAddr": link.YamiAddr(),
				}).Info("recycle link, yami dead")
				link.IsConnect = false
				link.yamiAddr = nil
			}

		}

}

func (y *Yume) AutoRecyle() {

	for {
		y.recycle()
		time.Sleep(1 * time.Minute)
	}
}

func (y *Yume) Start() {
	y.Tunnel.Start()
	go y.AutoRecyle()
}
