package sshspout

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strings"
)

const (
	maxMsgSize = 100
)

// TODO
type ComMsg struct {
	Host Host
	Msg  string
}

type Controller struct {
	hosts   HostConfig
	clients map[Host]*Session
	outChan chan ComMsg
}

func NewController(hc HostConfig) (*Controller, error) {
	clients := make(map[Host]*Session, len(hc))
	outChan := make(chan ComMsg, maxMsgSize)
	if err := hc.Check(); err != nil {
		return nil, err
	}
	return &Controller{clients: clients, hosts: hc, outChan: outChan}, nil
}

func (ctl *Controller) AddHost(h Host) error {
	return nil
}

func (ctl *Controller) DelHost(h Host) error {
	return nil
}

func (ctl Controller) Hosts() []Host {
	return ctl.hosts
}

func (ctl *Controller) Start() error {
	for _, h := range ctl.hosts {
		var config *ssh.ClientConfig
		if len(h.Pass) > 0 {
			config = &ssh.ClientConfig{
				User: h.User,
				Auth: []ssh.AuthMethod{
					ssh.Password("gpx"),
				},
			}
		} else {
			if strings.HasPrefix(h.Key, "~") {
				h.Key = os.Getenv("HOME") + h.Key[1:]
			}
			// Deal with memory leak
			privkey, err := ioutil.ReadFile(h.Key)
			if err != nil {
				return err
			}
			signer, err := ssh.ParsePrivateKey([]byte(privkey))
			if err != nil {
				return err
			}
			config = &ssh.ClientConfig{
				User: h.User,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
			}
		}

		client, err := ssh.Dial("tcp", h.Ip, config)
		if err != nil {
			return err
		}
		sess := NewSession(h, client)
		if err = sess.Start(ctl.outChan); err != nil {
			return err
		}
		ctl.clients[h] = sess
	}
	go GetResult(ctl.outChan)
	return nil
}

func (ctl *Controller) Wait() error {
	return nil
}

func (ctl *Controller) Close(h Host) error {
	if s, exist := ctl.clients[h]; exist {
		if err := s.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (ctl *Controller) CloseAll() error {
	hasErr := false
	for h := range ctl.clients {
		if err := ctl.Close(h); err != nil {
			hasErr = true
			log.Warnf("Close Host %s error: %v", h.Ip, err)
		}
	}
	if hasErr {
	}
	return nil
}

func (ctl *Controller) Run(cmd string) error {
	for _, s := range ctl.clients {
		s.In <- cmd
	}
	return nil
}

func GetResult(out chan ComMsg) {
	for {
		msg := <- out
		log.WithFields(log.Fields{
			"Type": "Recevice",
			"Host": msg.Host.Ip,
		}).Info(msg.Msg)
	}
}
