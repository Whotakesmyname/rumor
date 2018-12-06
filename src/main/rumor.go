package main

import (
	"io"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"service"

	"github.com/docopt/docopt-go"
)

// VERSION defines the version string showed in help message.
const VERSION = `0.1.0 alpha`

type config struct {
	Start     bool
	File      string
	Stop      bool
	Node      bool
	Self      bool
	Add       bool
	List      bool
	BucketIdx int `docopt:"<bucket-index>"`
	Ping      bool
	NodeStr   string `docopt:"<node-string>"`
	Update    bool
	NodeID    string `docopt:"<NodeID>"`
}

const usage = `Rumor.

Usage:
  rumor start [--file=<path/to/tree>]
  rumor stop
  rumor node self
  rumor node add <node-string>
  rumor node list <bucket-index>
  rumor node ping <node-string>
  rumor node update <NodeID>
  
Options:
  -h --help  Show this screen.
  --version  Show version.
  `

func cliHandler(conn net.Conn, server *service.Server) {
	defer func() {
		recover()
		conn.Write([]byte{0}) // Close conn.
		conn.Close()
		return
	}()
	var cfg config
	dec := gob.NewDecoder(conn)
	dec.Decode(&cfg)

	errHandler := func(err error) {
		if err != nil {
			conn.Write([]byte(err.Error()))
			panic(err)
		}
	}

	// Handle Req
	if cfg.Stop {
		log.Println("user requests to stop.")
		conn.Write([]byte{0}) // Success and close connection.
		conn.Close()
		os.Exit(0)
	} else if cfg.Node {
		if cfg.Add {
			var node service.Node
			err := node.DecodeString(cfg.NodeStr)
			errHandler(err)
			server.KBuckets.Add(node.ID, node.Address)
			log.Println("Successfully add offered new node.")
		} else if cfg.Self {
			conn.Write([]byte(server.KBuckets.Self.EncodeToString()))
		} else if cfg.List {
			bucket := server.KBuckets.Buckets[cfg.BucketIdx]
			if bucket == nil {
				conn.Write([]byte("Empty bucket."))
			} else {
				var buf bytes.Buffer
				idx := 0
				for ele := bucket.Queue.Front(); ele != nil; ele = ele.Next() {
					node := ele.Value.(*service.Node)
					fmt.Fprintf(&buf, "[%d]NodeID: %x\n   NodeString: %s\n", idx, *node.ID, node.EncodeToString())
					idx++
				}
				conn.Write(buf.Bytes())
			}
		} else if cfg.Ping {
			var node service.Node
			err := node.DecodeString(cfg.NodeStr)
			errHandler(err)
			conn.Write([]byte(fmt.Sprintf("Ping result: %t\n", server.Ping(&node))))
		} else if cfg.Update {
			var nodeID service.NodeID
			nodeIDSlice, err := hex.DecodeString(cfg.NodeStr)
			errHandler(err)
			copy(nodeID[:], nodeIDSlice)
			err := server.KBuckets.Update(&nodeID)
			if err != nil {
				conn.Write([]byte(err.Error()))
			} else {
				conn.Write([]byte("Node updated."))
			}
		}
		}
	}
	conn.Write([]byte{0}) // Success and close connection.
}

func initPrepare() {
	gob.Register(net.UDPAddr{})
}

func main() {
	var cfg config
	opts, _ := docopt.ParseArgs(usage, os.Args[1:], VERSION)
	err := opts.Bind(&cfg)
	if err != nil {
		panic(err)
	}

	// Server part
	if cfg.Start {
		initPrepare()
		var tree *service.BucketTree
		if cfg.File == "" {
			log.Println("Creating an empty bucket tree.")
			tree = service.NewBucketTree()
		} else {
			log.Println("Loading from an existing tree.")
			fd, err := os.Open(cfg.File)
			if err != nil {
				panic(err)
			}
			dec := gob.NewDecoder(fd)
			err = dec.Decode(&tree)
			if err != nil {
				panic(err)
			}
		}
		server := service.NewServer(tree)
		server.StartService()
		fmt.Printf("Rumor is running on local node:\nNodeID: %x\nAddress: %s\nNode String: %s\n", *server.KBuckets.Self.ID, server.KBuckets.Self.Address.String(), server.KBuckets.Self.EncodeToString())
		listener, err := service.NewNamedPipeListener()
		if err != nil {
			log.Panic(err)
		}
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("encountered an error when accepting an named pipe connection: %s\n", err)
				continue
			}
			go cliHandler(conn, server)
		}
	} else {
		// CLI part
		conn, err := service.DialPipe()
		if err != nil {
			log.Panic(err)
		}
		defer conn.Close()
		enc := gob.NewEncoder(conn)
		enc.Encode(cfg)
		buffer := make([]byte, 1024)
		for {
			n, err := conn.Read(buffer)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 1 && buffer[0] == 0 {
				return
			}
			fmt.Println(string(buffer[:n]))
		}
	}
}
