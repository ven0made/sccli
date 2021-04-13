package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("~~SCCLI~~")
	twitchChat("", "", "")
	vaushChat()
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

func vaushChat() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	type ChatMsgObj struct {
		Nick      string
		Features  []string
		Timestamp int
		Data      string
	}

	u := url.URL{Scheme: "wss", Host: "www.vaush.gg", Path: "/afori8vD4zjyfBjdmDSwLjrnytliIfSlVlxEW"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, _ := c.ReadMessage()

			messageType := strings.Split(string(message), " ")[0]
			messageJson := strings.Join(strings.Split(string(message), " ")[1:], " ")
			if messageType == "MSG" {
				var ChatMsg ChatMsgObj
				json.Unmarshal([]byte(messageJson), &ChatMsg)
				fmt.Println(ChatMsg.Nick+":", ChatMsg.Data)
			}

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		case <-interrupt:
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
