// +build !windows

package service

import "net"

const pipePath string = `/tmp/rumor.sock`

// NewNamedPipeListener creates a pipe listener on specific platform
// On unix, it is substituted with Unix Domain Socket.
func NewNamedPipeListener() (net.Listener, error) {
	return net.Listen("unix", pipePath)
}

// DialPipe dials a named pipe on specific platform.
// On unix, it is substituted with Unix Domain Socket.
func DialPipe() (net.Conn, error) {
	return net.Dial("unix", pipePath)
}
