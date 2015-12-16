package main

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"fmt"
	"net/http"
	"github.com/bitxel/sshspout/engine"
)

// CmdHandler handle the cmd from user
// send it to destination server and forward it to ui
// TODO:
//    get server configs from user.
func CmdHandler(ws *websocket.Conn) {
	log.Debugf("readWriteServer %#v\n", ws.Config())

	hosts := engine.HostConfig{
		engine.Host{IP: "vm:22", User: "xt", Key: "~/.ssh/id_rsa"},
		engine.Host{IP: "gpxtrade.com:22", User: "root", Key: "~/.ssh/id_rsa"},
	}
	ctl, err := engine.NewController(hosts)
	if err != nil {
		log.Errorf("create new controller error %v", err)
	}

	c, err := ctl.Start()
	if err != nil {
		log.Errorf("start session error %v:", err)
	}

	go func(out chan engine.Message){
		for {
			msg := <- out
			log.WithFields(log.Fields{
				"Type": msg.Type,
				"Host": msg.Host.IP,
			}).Info(msg.Msg)
			err = websocket.JSON.Send(ws, msg)
			if err != nil {
				log.Errorf("websocket send error:%v", err)
				break
			}
			log.Infof("send:%q\n", msg)
		}
	}(c)

	for {
		var buf string
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			log.Errorf("receive message error: %v", err)
			break
		}
		log.Infof("recv cmd:%q\n", buf)
		ctl.Run(buf+"\n")
	}
	ctl.CloseAll()
}

// main func of the whole application
func main() {
	log.SetLevel(log.InfoLevel)
	http.Handle("/cmd", websocket.Handler(CmdHandler))
	fmt.Println("Server started...")
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
