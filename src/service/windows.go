// +build windows

package service

import (
	"net"

	"github.com/Microsoft/go-winio"
)

const pipePath string = `\\.\pipe\RumorPipe`

// NewNamedPipeListener creates a pipe listener on specific platform
func NewNamedPipeListener() (net.Listener, error) {
	return winio.ListenPipe(pipePath, nil)
}

// DialPipe dials a named pipe on specific platform.
func DialPipe() (net.Conn, error) {
	return winio.DialPipe(pipePath, nil)
}
