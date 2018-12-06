// Package service offers miscellaneous tools
package service

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/bits"
	"net"
)

// NodeID is a node's hash id.
type NodeID [NodeIDLength]byte

// Cookie is a random hash used for defend fake responses.
type Cookie [CookieLength]byte

// NewRandNodeID returns a random NodeID. If failed, return nil.
func NewRandNodeID() *NodeID {
	var buffer NodeID
	_, err := rand.Read(buffer[:])
	if err != nil {
		return nil
	}
	return &buffer
}

// NewRandCookie returns a random cookie. If failed, return nil.
func NewRandCookie() *Cookie {
	var buffer Cookie
	_, err := rand.Read(buffer[:])
	if err != nil {
		return nil
	}
	return &buffer
}

// DumpUDPAddr dumps UDPAddr.
// Only IPv4 is supported yet.
func DumpUDPAddr(addr *net.UDPAddr) []byte {
	ip := addr.IP.To4()
	if ip == nil {
		return nil
	}
	// for future compatibility of IPv6
	port := make([]byte, 2)
	binary.LittleEndian.PutUint16(port, uint16(addr.Port))
	ip = append(ip, port...)
	return ip
}

// CommonPrefixLength calcs the length of common prefix bits of two nodeID slices.
// The two ID slices must share a same length, or -1 will be returned.
func CommonPrefixLength(a []byte, b []byte) int {
	length := len(a)
	if length != len(b) {
		return -1
	}
	count := 0
	for p, inc := 0, 0; p < length; p++ {
		inc = bits.LeadingZeros8(a[p] ^ b[p])
		count += inc
		if inc != 8 {
			break
		}
	}
	return count
}

// Min returns the smaller integer.
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// DetectPublicAddr try to find local machine's public udp address by STUN.
func DetectPublicAddr() net.Addr {
	return &net.UDPAddr{}
}

// WelcomePrint prints welcome message when rumor service is successfully initialed
func WelcomePrint() {
	fmt.Println("===================================================")
	fmt.Println("            Be Careful of Rumors")
	fmt.Println("===================================================")
}
