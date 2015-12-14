package sshspout_test

import (
	"github.com/bitxel/sshspout"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	hosts := sshspout.HostConfig{
		sshspout.Host{IP: "vm:22", User: "xt", Key: "~/.ssh/id_rsa"},
		sshspout.Host{IP: "gpxtrade.com:22", User: "root", Key: "~/.ssh/id_rsa"},
	}
	ctl, err := sshspout.NewController(hosts)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ctl.Hosts())
	if err = ctl.Start(); err != nil {
		t.Fatal(err)
	}
	ctl.Run("uptime\n")
	ctl.Run("whoami\n")
	ctl.Run("exit\n")
	time.Sleep(time.Second * 5)
	ctl.CloseAll()
}
