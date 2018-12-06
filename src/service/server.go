package service

import (
	"fmt"
	"log"
	"net"
)

// Server struct used for communication
type Server struct {
	CookieTable Table
	KBuckets    *BucketTree
	conn        net.PacketConn
	stop        bool
}

// Protocol is an interface defines all possible types of communication.
type Protocol interface {
	Ping(*Node) bool
}

// NewServer creates a server
// It must load from an existing K-Bucket tree instance.
func NewServer(tree *BucketTree) *Server {
	if tree == nil {
		return nil
	}
	// Listening local port.
	conn, err := net.ListenPacket("udp", fmt.Sprintf(`:%d`, Port))
	if err != nil {
		return nil
	}
	return tree.SetServerInstance(&Server{NewCookieTable(), tree, conn, false})
}

// StartService starts the message handler loop.
// This function deals with recognizing incoming data type and distributing to other handlers.
func (server *Server) StartService() {
	// Start response & request handler
	responseChan := make(chan *Datagram, ResponseHandlerQueueLength)
	requestChan := make(chan *Datagram, RequestHandlerQueueLength)
	go server.responseHandler(responseChan)
	go server.requestHandler(requestChan)

	// Incoming messages detection and distribution loop
	go func() {
		var buffer [MaxPackageSize]byte
		var datagram *Datagram
		defer server.conn.Close()
		for {
			if server.stop {
				break // STOP
			}
			n, addr, err := server.conn.ReadFrom(buffer[:])
			// Any IO error or length less than minimal possible length will be abandoned.
			if err != nil || n < (CookieLength+NodeIDLength+9) {
				continue
			}
			datagram = new(Datagram).Loads(buffer[:n], addr)

			// Welcome every node except the msg is a pong response
			// Place welcome here because I simply don't want to pass argument `addr` to upper layer.
			if datagram.Type != Ping || datagram.IsRequest {
				go server.welcomeNode(datagram)
			}

			if datagram.IsRequest {
				requestChan <- datagram
				continue
			} else {
				// If incoming message is a response to a former request from self
				responseChan <- datagram
				continue
			}
		}
	}()
	WelcomePrint()
}

// Stop stops the server.
func (server *Server) Stop() {
	server.stop = true
}

// Response handler
// Route responses causing by former requests from local.
func (server *Server) responseHandler(inChan <-chan *Datagram) {
	var datagram *Datagram
	for {
		datagram = <-inChan
		source := server.CookieTable.Get(datagram.MagicCookie)
		if source == nil {
			continue
		}
		source <- datagram
	}
}

// Request handler
// Reply to incoming requests.
func (server *Server) requestHandler(outChan <-chan *Datagram) {
	for {
		datagram := <-outChan
		switch datagram.Type {
		case Ping:
			go server.rePing(datagram)
			break
		}
	}
}

// Welcome a new node or update a existing node.
// Simple helper method.
func (server *Server) welcomeNode(datagram *Datagram) {
	server.KBuckets.Add(datagram.SourceNode.ID, datagram.SourceNode.Address)
}

/*
Logic:
For requests, each request should give a destination node and extra content if any. A channel will be returned for the response in the future.
Cookies are used for distinguishing incoming data, whether they are meant to request or are responses of a former request.

*/

// Ping implementation.
// This method cannot attach Ping to a RPC reply.
func (server *Server) Ping(node *Node) bool {
	cookie := NewRandCookie()
	if cookie == nil {
		return false
	}
	ptrDatagram := NewDatagram(Ping, true, cookie, server.KBuckets.Self, NewPing())
	if ptrDatagram == nil {
		return false
	}
	resChan := make(chan *Datagram, 1)
	if server.CookieTable.Add(cookie, resChan) == nil {
		return false
	}

	_, err := server.conn.WriteTo(ptrDatagram.Dumps(), node.Address)
	if err != nil {
		return false
	}
	// Wait for response
	_, ok := <-resChan
	if ok {
		return true
	}
	return false
}

// response Ping request.
func (server *Server) rePing(datagram *Datagram) {
	resDatagram := NewDatagram(Ping, false, datagram.MagicCookie, server.KBuckets.Self, NewPing())
	server.conn.WriteTo(resDatagram.Dumps(), datagram.SourceNode.Address)
	log.Println("A ping response has been sent out.")
}
