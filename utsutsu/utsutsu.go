package main

import "github.com/sirupsen/logrus"

type Utsutsu struct {
	logger    *logrus.Logger
	Config    map[string]interface{}

	Tunnel *Tunnel
	Cmder *Cmder
	Hikari *Hikari
	Yami *Yami
	Ui *Ui

	addr string
}

func NewUtsutsu() *Utsutsu {

	u := new(Utsutsu)

	u.logger = logrus.New()

	u.Config = make(map[string]interface{})
	//u.Config["yumeKcpAddr"] = "shitama.sakuya.love:31338"
	u.Config["yumeKcpAddr"] = "127.0.0.1:31338"
	u.Config["yumeKcpAddrAlt"] = "39.106.32.93:31338"

	u.Tunnel = NewTunnel(u)
	u.Cmder = NewCmder(u)
	u.Hikari = NewHikari(u)
	u.Yami = NewYami(u)



	u.Tunnel.OnConnected.SubscribeAsync("", func() {

		u.logger.WithFields(logrus.Fields{
			"scope": "utsutsu/handleConnected",
		}).Info("connected")

	}, false)

	u.Tunnel.OnDisconnected.SubscribeAsync("", func() {

		u.logger.WithFields(logrus.Fields{
			"scope": "utsutsu/handleDisconnected",
		}).Info("disconnected")

	}, false)

	return u

}

func (u *Utsutsu) Start() {
	u.Tunnel.Start()
	u.Cmder.Start()
}