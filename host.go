package sshspout

import (
	"errors"
)

var (
	ErrIpNull         = errors.New("Host Ip is null")
	ErrUserNull       = errors.New("Host User is null")
	ErrPassAndKeyNull = errors.New("Host password and private key both null")
)

type Host struct {
	Ip   string
	User string
	Pass string
	Key  string
}

type HostConfig []Host

func (hc HostConfig) Check() error {
	for _, v := range hc {
		if len(v.Ip) == 0 {
			return ErrIpNull
		}
		if len(v.User) == 0 {
			return ErrUserNull
		}
		if len(v.Pass) == 0 && len(v.Key) == 0 {
			return ErrPassAndKeyNull
		}
	}
	return nil
}
