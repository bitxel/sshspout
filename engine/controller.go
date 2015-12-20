package engine

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strings"
)

// Controller is the core engine to deal with multiple connections
type Controller struct {
	hosts   map[HostID]Host
	clients map[HostID]*Session
	outChan chan Message
}

// NewController init a new Controller
func NewController(count int) *Controller {
	clients := make(map[HostID]*Session, count)
	hosts := make(map[HostID]Host, count)
	outChan := make(chan Message, maxMsgSize)
	return &Controller{clients: clients, hosts: hosts, outChan: outChan}
}

// AddHost Add a server
func (ctl *Controller) AddHost(hid HostID, h Host) error {
	if err := h.Check(); err != nil {
		return err
	}
	ctl.hosts[hid] = h
	return nil
}

// DelHost Del a server
func (ctl *Controller) DelHost(hid HostID) error {
	if _, exist := ctl.hosts[hid]; exist {
		// TODO close server conn before delete
		delete(ctl.hosts, hid)
	}
	return nil
}

// Hosts return the host list
func (ctl Controller) Hosts() map[HostID]Host {
	hosts := make(map[HostID]Host, len(ctl.hosts))
	for k, v := range ctl.hosts {
		hosts[k] = v
	}
	return hosts
}

// Start to connect to the host, init the input and output channel
// TODO return error with specific host
func (ctl *Controller) Start() (chan Message, error) {
	for hid, h := range ctl.hosts {
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
				return nil, err
			}
			signer, err := ssh.ParsePrivateKey([]byte(privkey))
			if err != nil {
				return nil, err
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
			return nil, err
		}
		sess := NewSession(hid, h, client)
		if err = sess.Start(ctl.outChan); err != nil {
			return nil, err
		}
		ctl.clients[hid] = sess
	}
	//go GetResult(ctl.outChan)
	return ctl.outChan, nil
}

// Wait for actions done
// TODO
func (ctl *Controller) Wait() error {
	return nil
}

// Close a single host
func (ctl *Controller) Close(hid HostID) error {
	if s, exist := ctl.clients[hid]; exist {
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
			log.Warnf("Close Host %s error: %v", ctl.hosts[h].IP, err)
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
