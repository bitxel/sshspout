package engine

import (
	"errors"
)

var (
	// ErrIPNull is the error IP is null in the config
	ErrIPNull = errors.New("Host Ip is null")
	// ErrUserNull is the error User is null in the config
	ErrUserNull       = errors.New("Host User is null")
	// ErrPassAndKeyNull is the error Password or Public Key are both null
	ErrPassAndKeyNull = errors.New("Host password and private key both null")
)

// Host is the basic info of a server
type Host struct {
	IP string
	User string
	Pass string
	Key  string
}

// Check is to validate the host config
func (h Host) Check() error {
	if len(h.IP) == 0 {
		return ErrIPNull
	}
	if len(h.User) == 0 {
		return ErrUserNull
	}
	if len(h.Pass) == 0 && len(h.Key) == 0 {
		return ErrPassAndKeyNull
	}
	return nil
}
