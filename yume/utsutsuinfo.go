package main

import (
	"github.com/hashicorp/yamux"
	kcp "github.com/xtaci/kcp-go"
)

type UtsutsuInfo struct {
	Addr string `json:"addr"`
	IP   string `json:"ip"`
	// EchoPort int    `json:"echoPort"`
	// Hoster   bool   `json:"hoster"`
	// NickName string `json:"nickname"`

	conn    *kcp.UDPSession
	session *yamux.Session
}

func NewUtsutsuInfo(addr string, ip string, conn *kcp.UDPSession, session *yamux.Session) UtsutsuInfo {

	u := UtsutsuInfo{Addr: addr, IP: ip}
	u.conn = conn
	u.session = session

	return u

}
