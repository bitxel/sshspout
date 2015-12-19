package engine_test

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitxel/sshspout/engine"
	"testing"
	"time"
)

// GetResult  get the output and print to the screen
func GetResult(out chan engine.Message) {
	for {
		msg := <- out
		log.WithFields(log.Fields{
			"Type": msg.Type,
			"Host": msg.Host.IP,
		}).Info(msg.Msg)
	}
}
func TestController(t *testing.T) {

	hosts := []engine.Host{
		engine.Host{IP: "vm:22", User: "xt", Key: "~/.ssh/id_rsa"},
		engine.Host{IP: "gpxtrade.com:22", User: "root", Key: "~/.ssh/id_rsa"},
	}
	ctl := engine.NewController(len(hosts))
	for k, v := range hosts {
		if err:= v.Check(); err != nil {
			t.Fatal(err)
		}
		ctl.AddHost(engine.HostID(k),v)
	}
	t.Log(ctl.Hosts())
	c, err := ctl.Start()
	if err != nil {
		t.Fatal(err)
	}
	go GetResult(c)
	ctl.Run("uptime\n")
	ctl.Run("whoami\n")
	ctl.Run("exit\n")
	time.Sleep(time.Second * 5)
	ctl.CloseAll()
}
