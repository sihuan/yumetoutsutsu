package main

import "fmt"

type Cmder struct {
	parent *Utsutsu
}

func NewCmder(parent *Utsutsu) *Cmder {
	c := new(Cmder)
	c.parent = parent

	return c
}

func (c *Cmder) Start() {
	var cmd string

	go func() {
		for {
			cmd = ""
			fmt.Print("Utsutsu--> ")
			fmt.Scanln(&cmd)
			switch cmd {
			case "h":
				c.help()
				break
			case "x":
				c.connect()
				break
			case "c":
				c.create()
			}
		}
	}()
}

func (c *Cmder) help() {
	fmt.Println("输入 'x' 连接别人，输入 'c' 建立主机")
}

func (c *Cmder) create() {

	nickName := ""
	for {
		fmt.Println("输入昵称，纯英文别太长")
		fmt.Scanln(&nickName)
		if nickName != "" {
			break
		}
	}
	c.parent.Hikari.nickName = nickName

	c.parent.Hikari.Start()

	exit := ""
	for {
		fmt.Println("输入 q 回车退出")
		fmt.Scanln(&exit)
		if exit == "q" {
			c.parent.Hikari.Stop()
			break
		}
	}
}

func (c *Cmder) connect() {
	c.parent.Yami.Start()
}
