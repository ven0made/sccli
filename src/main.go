package main

import (
	"crypto/tls"
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

func main() {
	fmt.Println("~~SCCLI~~")
	twitchChat("", "", "")
}

func twitchChat(username, oauth, channel string) {

	ircCfg := irc.NewConfig(username)
	ircCfg.SSL = true
	ircCfg.SSLConfig = &tls.Config{ServerName: "irc.chat.twitch.tv"}
	ircCfg.Server = "irc.chat.twitch.tv:443"
	ircCfg.Pass = oauth

	ircClient := irc.Client(ircCfg)

	ircClient.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		conn.Join("#" + channel)
	})

	quit := make(chan bool)
	ircClient.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})

	ircClient.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(line.Nick + ": " + line.Args[1])
	})

	ircClient.Connect()

	<-quit
}
