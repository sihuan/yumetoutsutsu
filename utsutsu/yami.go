package main

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type Yami struct {
	parent *Utsutsu
	links  []Link

	pcGame   net.PacketConn
	pcYume   net.PacketConn
	YumeAddr net.Addr
	GameAddr net.Addr
}

func NewYami(parent *Utsutsu) *Yami {
	y := new(Yami)
	y.parent = parent

	y.links = make([]Link, 0)

	return y
}

func (y *Yami) Start() {

	y.links = y.parent.Tunnel.zGetLinks()

	// fmt.Print(y.links)
	if len(y.links)==0 {
		fmt.Println("没有已经建立的主机。。")
		return
	}
	for i, link := range y.links {
		fmt.Printf("%d    %s     ", i, link.NickName)
		if link.IsConnect == true {
			fmt.Print("*")
		}
		fmt.Print("\n")
	}

	num := -1

	for {
		fmt.Println("输入连接序号")
		fmt.Scanln(&num)
		if num >= 0 && num < len(y.links) {
			break
		}
	}

	y.YumeAddr, _ = net.ResolveUDPAddr("udp4", y.links[num].ToYamiAddr)

	pcGame, err := net.ListenPacket("udp4", "127.0.0.1:31106")

	if err != nil {
		y.parent.logger.WithFields(logrus.Fields{
			"scope": "yami/Start",
		}).Fatal(err)
	}

	y.pcGame = pcGame

	pcYume, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		y.parent.logger.WithFields(logrus.Fields{
			"scope": "yami/Start",
		}).Fatal(err)
	}

	y.pcYume = pcYume

	go y.keepAlive()
	go y.handleGameConnection()
	go y.handleYumeConnection()

	fmt.Println("连接 127.0.0.1:31106 即可")
	exit := ""
	for {
		fmt.Println("输入 q 回车退出")
		fmt.Scanln(&exit)
		if exit == "q" {
			y.Stop()
			break
		}
	}

}

func (y *Yami) Stop() {
	y.GameAddr = nil
	y.pcGame.Close()
	y.pcYume.Close()
}

func (y *Yami) keepAlive() {
	for {

		_, err := y.pcYume.WriteTo([]byte("PHANTOM"), y.YumeAddr)

		if err != nil {
			y.parent.logger.WithFields(logrus.Fields{
				"scope": "yami/keepAlive",
			}).Warn(err)
			break
		}

		time.Sleep(10 * time.Second)

	}
}

func (y *Yami) handleGameConnection() {

	buf := make([]byte, 1536)

	for {

		n, addr, err := y.pcGame.ReadFrom(buf)

		fmt.Println(buf[:3])

		if err != nil {
			//y.parent.logger.WithFields(logrus.Fields{
			//	"scope": "yami/handleGameConnection",
			//}).Warn(err)
			break
		}

		if y.GameAddr == nil {
			y.GameAddr = addr
		}
		// fmt.Println("G:", buf[:3], y.pcYume.LocalAddr().String(), "-->", y.YumeAddr)

		y.pcYume.WriteTo(buf[:n], y.YumeAddr)
	}
}

func (y *Yami) handleYumeConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := y.pcYume.ReadFrom(buf)
		// fmt.Println("Y:", buf[:3])

		if err != nil {
			//y.parent.logger.WithFields(logrus.Fields{
			//	"scope": "yami/handleYumeConnection",
			//}).Warn(err)
			break
		}
		y.pcGame.WriteTo(buf[:n], y.GameAddr)
	}
}
