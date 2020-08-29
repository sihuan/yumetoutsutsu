package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Link struct {
	ToYamiAddr string
	IsConnect  bool
	NickName   string

	utsutsuAddr string

	parent     *Yume
	pcHikari   net.PacketConn
	pcYami     net.PacketConn
	hikariAddr net.Addr
	yamiAddr   net.Addr

	hikariActive time.Time
	yamiActive   time.Time
}

func (y *Yume) NewLink(utsutsuAddr string, nickName string) *Link {

	l := new(Link)
	l.parent = y
	l.utsutsuAddr = utsutsuAddr
	l.IsConnect = false
	l.NickName = nickName

	l.Start()

	return l

}

func (l *Link) Start() {

	pcHikari, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.logger.WithFields(logrus.Fields{
			"scope": "link/Start",
		}).Fatal(err)
	}

	l.pcHikari = pcHikari

	pcYami, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.logger.WithFields(logrus.Fields{
			"scope": "link/Start",
		}).Fatal(err)
	}

	l.pcYami = pcYami
	l.ToYamiAddr = l.YamiAddr()

	l.hikariActive = time.Now()

	go l.handleHikariConnection()
	go l.handleYamiConnection()

}

func (l *Link) handleHikariConnection() {

	buf := make([]byte, 1536)

	for {

		n, addr, err := l.pcHikari.ReadFrom(buf)

		if err != nil {
			l.parent.logger.WithFields(logrus.Fields{
				"scope": "link/handleHikariConnection",
			}).Warn(err)
			break
		}

		l.hikariActive = time.Now()

		if bytes.Equal(buf[:n], []byte("PHANTOM")) {
			if l.hikariAddr == nil {
				l.hikariAddr = addr
				l.parent.logger.WithFields(logrus.Fields{
					"scope": "link/handleHikariConnection",
				}).Info("hikari bound")
			}
		} else {
			l.pcYami.WriteTo(buf[:n], l.yamiAddr)
		}

	}

}

func (l *Link) handleYamiConnection() {

	buf := make([]byte, 1536)

	for {

		n, addr, err := l.pcYami.ReadFrom(buf)

		if err != nil {
			l.parent.logger.WithFields(logrus.Fields{
				"scope": "link/handleYamiConnection",
			}).Warn(err)
			break
		}

		l.yamiActive = time.Now()

		if bytes.Equal(buf[:n], []byte("PHANTOM")) {
			if l.yamiAddr == nil {
				l.yamiAddr = addr
				l.parent.logger.WithFields(logrus.Fields{
					"scope": "link/handleYamiConnection",
				}).Info("Yami bound")
				l.IsConnect = true
			}
		} else {
			l.pcHikari.WriteTo(buf[:n], l.hikariAddr)
		}

	}

}

func (l *Link) HikariAddr() string {
	return fmt.Sprintf("%s:%s",
		l.parent.PublicAddr,
		strings.Split(l.pcHikari.LocalAddr().String(), ":")[1],
	)
}

func (l *Link) YamiAddr() string {
	return fmt.Sprintf("%s:%s",
		l.parent.PublicAddr,
		strings.Split(l.pcYami.LocalAddr().String(), ":")[1],
	)
}

func (l *Link) Stop() {

	l.pcHikari.Close()
	l.pcYami.Close()

}

func (l *Link) HikariExpired() bool {
	return time.Now().Sub(l.hikariActive).Seconds() > 31
}

func (l *Link) YamiExpired() bool {
	return time.Now().Sub(l.yamiActive).Seconds() > 31
}
