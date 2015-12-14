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

// MsgType basic message type
type MsgType int

const (
	_ MsgType = iota
	// MsgOpen open connection msg
	MsgOpen
	// MsgSend msg to server
	MsgSend
	// MsgReceived msg from server
	MsgReceived
	// MsgClose close msg
	MsgClose
	// MsgError error msg
	MsgError
)

// Message between sever and client
type Message struct {
	Type MsgType
	Host Host
	Msg  string
}

// Controller is the core engine to deal with multiple connections
type Controller struct {
	hosts   HostConfig
	clients map[Host]*Session
	outChan chan Message
}

// NewController init a new Controller
func NewController(hc HostConfig) (*Controller, error) {
	clients := make(map[Host]*Session, len(hc))
	outChan := make(chan Message, maxMsgSize)
	if err := hc.Check(); err != nil {
		return nil, err
	}
	return &Controller{clients: clients, hosts: hc, outChan: outChan}, nil
}

// AddHost Add a server
// TODO
func (ctl *Controller) AddHost(h Host) error {
	return nil
}

// DelHost Del a server
// TODO
func (ctl *Controller) DelHost(h Host) error {
	return nil
}

// Hosts return the host list
func (ctl Controller) Hosts() []Host {
	return ctl.hosts
}

// Start to connect to the host, init the input and output channel
func (ctl *Controller) Start() error {
	for _, h := range ctl.hosts {
		var config *ssh.ClientConfig
		if len(h.Pass) > 0 {
			config = &ssh.ClientConfig{
				User: h.User,
				Auth: []ssh.AuthMethod{
					ssh.Password(h.Pass),
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

		client, err := ssh.Dial("tcp", h.IP, config)
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

// Wait for actions done
// TODO
func (ctl *Controller) Wait() error {
	return nil
}

// Close a single host
func (ctl *Controller) Close(h Host) error {
	if s, exist := ctl.clients[h]; exist {
		if err := s.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CloseAll connections and clean the spot
func (ctl *Controller) CloseAll() error {
	hasErr := false
	for h := range ctl.clients {
		if err := ctl.Close(h); err != nil {
			hasErr = true
			log.Warnf("Close Host %s error: %v", h.IP, err)
		}
	}
	// Return to Frontend
	if hasErr {
	}
	return nil
}

// Run a command from user
// TODO Run cmd on specific hosts
func (ctl *Controller) Run(cmd string) error {
	for _, s := range ctl.clients {
		s.In <- cmd
	}
	return nil
}

// GetResult  get the output and print to the screen
func GetResult(out chan Message) {
	for {
		msg := <- out
		log.WithFields(log.Fields{
			"Type": msg.Type,
			"Host": msg.Host.IP,
		}).Info(msg.Msg)
	}
}
