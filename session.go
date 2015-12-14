package sshspout

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
)

// Session contains basic info of the conn, and the input and output chan
type Session struct {
	Host    Host
	client  *ssh.Client
	session *ssh.Session
	In      chan string
	Out     chan string
	Exit    chan bool
}

// NewSession start a session
func NewSession(host Host, client *ssh.Client) *Session {
	in := make(chan string, 100)
	out := make(chan string, 100)
	exit := make(chan bool, 1)
	return &Session{Host: host, client: client, In: in, Out: out, Exit: exit}
}

func getStream(s *ssh.Session) (in io.WriteCloser, out io.Reader, stderr io.Reader, err error) {
	in, err = s.StdinPipe()
	if err != nil {
		return
	}
	out, err = s.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err = s.StderrPipe()
	if err != nil {
		return
	}
	return
}

func streamToChan(source io.Reader, host Host, out chan <- Message) {
	for {
		buf := make([]byte, 1<<10)
		n, err := source.Read(buf)
		if n > 0 {
			out <- Message{Type: MsgReceived, Host: host, Msg:string(buf[0:n])}
		}
		if err == io.EOF {
			out <- Message{Type: MsgClose, Host: host}
			return
		}
		if err != nil {
			log.Fatalf("error read from stream: %v", err)
		}
	}
}

// Start initialize a connection to server
func (s *Session) Start(out chan Message) error {
	sess, err := s.client.NewSession()
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0, // disable echoing
		ssh.VREPRINT:      0,
		ssh.CS8:           1,
		ssh.ECHOE:         1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := sess.RequestPty("xterm", 80, 140, modes); err != nil {
		return err
	}

	stdin, stdout, stderr, err := getStream(sess)
	if err != nil {
		return err
	}
	go streamToChan(stdout, s.Host, out)
	go streamToChan(stderr, s.Host, out)
	go s.exeCmd(stdin)
	err = sess.Shell()
	if err != nil {
		log.Println("generate shell err: ", err)
	}

	log.WithFields(log.Fields{
		"Type": "Status",
		"Host": s.Host.IP,
	}).Info("Session Start")
	return nil
}

func (s *Session) exeCmd(in io.WriteCloser) {
	for {
		cmd := <-s.In
		if len(cmd) > 0 {
			log.WithFields(log.Fields{
				"Type": "Command",
				"Host": s.Host.IP,
			}).Info(cmd)
			in.Write([]byte(cmd))
		}
	}
}
