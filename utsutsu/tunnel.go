package main

import (
	"time"

	"encoding/gob"

	"github.com/asaskevich/EventBus"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Tunnel struct {
	OnConnected       EventBus.Bus
	OnDisconnected    EventBus.Bus
	OnRequestReceived EventBus.Bus
	OnPublishReceived EventBus.Bus
	OnPipeReceived    EventBus.Bus
	parent            *Utsutsu
	conn              *kcp.UDPSession
	session           *yamux.Session
}

func NewTunnel(parent *Utsutsu) *Tunnel {

	t := new(Tunnel)
	t.parent = parent

	t.OnConnected = EventBus.New()
	t.OnDisconnected = EventBus.New()
	t.OnRequestReceived = EventBus.New()
	t.OnPublishReceived = EventBus.New()
	t.OnPipeReceived = EventBus.New()

	t.OnDisconnected.SubscribeAsync("", func() {
		go t.Start()
	}, true)

	return t

}

func (t *Tunnel) Start() {

	yumeKcpAddr := t.parent.Config["yumeKcpAddr"].(string)
	yumeKcpAddrAlt := t.parent.Config["yumeKcpAddrAlt"].(string)

	var err error
	var conn *kcp.UDPSession

	conn, err = kcp.DialWithOptions(yumeKcpAddr, nil, 10, 3)

	if err != nil {

		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/Start",
		}).Warn(err)

		conn, err = kcp.DialWithOptions(yumeKcpAddrAlt, nil, 10, 3)

		if err != nil {
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/Start",
			}).Fatal(err)
		}

	}

	t.conn = conn

	go t.handleConnection(conn)

}

func (t *Tunnel) handleConnection(conn *kcp.UDPSession) {

	config := yamux.DefaultConfig()
	config.LogOutput = t.parent.logger.WithFields(logrus.Fields{
		"scope": "tunnel/yamux",
	}).Writer()

	session, err := yamux.Client(conn, config)

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/handleConnection",
		}).Fatal(err)
	}

	t.session = session

	_, err = t.session.Ping()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/handleConnection",
		}).Warn(err)
		t.stopConnection()
		return
	}

	go t.keepAlive()

	t.zInitUtsutsu()

	t.OnConnected.Publish("")

}

func (t *Tunnel) stopConnection() {

	err := t.session.GoAway()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	err = t.session.Close()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	err = t.conn.Close()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	t.OnDisconnected.Publish("")

}

func (t *Tunnel) keepAlive() {

	for {

		_, err := t.session.Ping()

		if err != nil {
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/keepAlive",
			}).Warn(err)
			t.stopConnection()
			break
		}

		//log.Printf("rtt is %dms", duration/time.Millisecond)

		time.Sleep(1 * time.Second)

	}

}

func (t *Tunnel) zInitUtsutsu() {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zInitUtsutus",
		}).Warn(err)
		return
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("utsutsu.init")

	var addr string

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&addr)

	t.parent.addr = addr

	stream.Close()

}

func (t *Tunnel) zGetLinks() []Link {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zGetLinks",
		}).Warn(err)
		return nil
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("links")

	data := make([]Link, 0)

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&data)

	stream.Close()

	return data

}

func (t *Tunnel) zNewLink() (rHikariAddr string) {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zNewLink",
		}).Fatal(err)
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("nlink")
	encoder.Encode(t.parent.addr)
	encoder.Encode(t.parent.Hikari.nickName)

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&rHikariAddr)

	stream.Close()

	return rHikariAddr

}

func (t *Tunnel) zDeleteLink() {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zNewLink",
		}).Fatal(err)
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("dlink")
	encoder.Encode(t.parent.addr)

	decoder := gob.NewDecoder(stream)
	var idx int
	decoder.Decode(&idx)

}
