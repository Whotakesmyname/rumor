package service

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// Node defines a node containing necessary info about another node
type Node struct {
	ID *NodeID
	// All nodes' address should be its public address for connection.
	Address net.Addr
}

func (node *Node) String() string {
	return fmt.Sprintf("Node %x at %s", *node.ID, node.Address.String())
}

// DecodeString creates a node from a base64 string
func (node *Node) DecodeString(str string) error {
	byteArr, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	// The byte array format:
	// | NodeID 20bytes | IPv4 4bytes | port 2bytes|
	if len(byteArr) != NodeIDLength+4+2 {
		return errors.New("illegal code string")
	}
	p := 0
	var nodeID NodeID
	copy(nodeID[:], byteArr[p:p+NodeIDLength])
	p += NodeIDLength
	node.ID = &nodeID

	ip := net.IPv4(byteArr[p], byteArr[p+1], byteArr[p+2], byteArr[p+3])
	p += 4
	port := int(binary.LittleEndian.Uint16(byteArr[p:]))
	node.Address = &net.UDPAddr{IP: ip, Port: port}
	return nil
}

// EncodeToString returns the string representation of the node
func (node *Node) EncodeToString() string {
	var buffer bytes.Buffer
	buffer.Write((*node.ID)[:])
	ip := node.Address.(*net.UDPAddr).IP.To4()
	if ip == nil {
		panic("error-the node has illegal ip")
	}
	buffer.Write(ip)
	port := make([]byte, 2)
	binary.LittleEndian.PutUint16(port, uint16(node.Address.(*net.UDPAddr).Port))
	buffer.Write(port)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}
