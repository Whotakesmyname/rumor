package service

import (
	"encoding/binary"
	"net"
	"time"
)

// Define request type
const (
	Request byte = 0x80 // 0b10000000 used for set flag on Type to distinguish request or response

	Ping byte = iota
)

// Datagram defines the datagram structure which is used for transmission
type Datagram struct {
	Type        byte // Type's highest bit has been resolved.
	IsRequest   bool // Claims itself as a request(true)/response(false).
	MagicCookie *Cookie
	SourceNode  *Node
	Timestamp   uint64
	Payload     []byte
}

// Payload defines different protocols' payload
type Payload interface {
	Dump() []byte
}

// DataPing is ping payload
type DataPing struct {
	data []byte
}

// NewPing creates ping payload.
// Here differs from the paper, ping is not considered to be attached in a RPC reply. However, this could be implemented in the future if necessary.
func NewPing() *DataPing {
	return &DataPing{[]byte{}}
}

// Dump dumps the payload to byte slice for transmission.
func (ping *DataPing) Dump() []byte {
	return ping.data
}

// NewDatagram creates a datagram.
// When used for reply, cookie should be passed; otherwise it could be nil to auto-generate.
func NewDatagram(msgType byte, isReq bool, cookie *Cookie, sourceNode *Node, payload Payload) *Datagram {
	if cookie == nil {
		cookie = NewRandCookie()
		if cookie == nil {
			return nil
		}
	}
	timestamp := uint64(time.Now().UnixNano())
	payloadBytes := payload.Dump()
	if (NodeIDLength + CookieLength + len(payloadBytes) + 11) > MaxPackageSize {
		return nil
	}
	return &Datagram{msgType, isReq, cookie, sourceNode, timestamp, payloadBytes}
}

// Loads loads a datagram from byte slice and net.Addr
// All sources are a copy of their original ones for detaching from original buffer.
func (datagram *Datagram) Loads(bytes []byte, addr net.Addr) *Datagram {
	if len(bytes) < CookieLength+NodeIDLength+9 {
		return nil
	}
	p := 0
	datagram.Type = bytes[p] & ^Request
	datagram.IsRequest = bytes[p]&Request == Request
	p++
	cookie := new(Cookie)
	copy((*cookie)[:], bytes[p:p+CookieLength])
	datagram.MagicCookie = cookie
	p += CookieLength
	id := new(NodeID)
	copy((*id)[:], bytes[p:p+NodeIDLength])
	datagram.SourceNode = &Node{id, addr}
	p += NodeIDLength
	datagram.Timestamp = binary.LittleEndian.Uint64(bytes[p : p+8])
	p += 8
	payload := make([]byte, len(bytes)-p)
	copy(payload, bytes[p:])
	datagram.Payload = payload
	return datagram
}

// Dumps dumps data to []byte for transmission. Parenthesis values are default.
// |     Type     |    Cookie    | NodeID | Timestamp | Payload |
// |   1 byte(s)  |      20      |   20   |     8     |   ...   |
func (datagram *Datagram) Dumps() []byte {
	totalLength := NodeIDLength + CookieLength + len(datagram.Payload) + 11
	buffer := make([]byte, totalLength)

	p := 0
	if datagram.IsRequest {
		buffer[p] = datagram.Type | Request
	} else {
		buffer[p] = datagram.Type
	}
	p++
	copy(buffer[p:], (*datagram.MagicCookie)[:])
	p += CookieLength
	copy(buffer[p:], (*datagram.SourceNode.ID)[:])
	p += NodeIDLength
	binary.LittleEndian.PutUint64(buffer[p:], datagram.Timestamp)
	p += 8
	copy(buffer[p:], datagram.Payload)
	return buffer
}
