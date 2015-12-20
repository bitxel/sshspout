package main

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"fmt"
	"net/http"
	"github.com/bitxel/sshspout/engine"
	"encoding/json"
	"strconv"
)

// CmdHandler handle the cmd from user
// send it to destination server and forward it to ui
// TODO:
//    get server configs from user.
func CmdHandler(ws *websocket.Conn) {
	log.Debugf("readWriteServer %#v\n", ws.Config())

	hosts := []engine.Host{
		engine.Host{IP: "vm:22", User: "xt", Key: "~/.ssh/id_rsa"},
		engine.Host{IP: "gpxtrade.com:22", User: "root", Key: "~/.ssh/id_rsa"},
	}
	ctl := engine.NewController(len(hosts))
	for k, v := range hosts {
		if err:=v.Check(); err!= nil {
			log.Errorf("server conf err: %s", v)
		}
		ctl.AddHost(engine.HostID(strconv.Itoa(k)), v)
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
				"Host": msg.HostID,
			}).Info(msg.Msg)
			res, _ := json.Marshal(msg)
			log.Println(string(res))
			err = websocket.Message.Send(ws, string(res))
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
	http.Handle("/", http.FileServer(http.Dir("./ui")))
	fmt.Println("Server started...")
	err := http.ListenAndServe("127.0.0.1:9999", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
