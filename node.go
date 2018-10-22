package gossipnet

import (
	"bufio"
	"log"
	"net"
)

// The local Node
type Node struct {
	localAddr   string
	remoteAddrs []string
	ln          net.Listener
	running     bool
	remoteNodes map[string]*remoteNode
	OutChan     chan *Message
}

// Definition of a remote Node
type remoteNode struct {
	Data interface{}
	Conn net.Conn
}

// Create a new Network Gossip Node
func New(publicKey, localAddr string, remoteAddrs []string) (*Node, error) {
	return &Node{
		localAddr:   localAddr,
		remoteAddrs: remoteAddrs,
		running:     false,
		remoteNodes: make(map[string]*remoteNode),
		OutChan:     make(chan *Message),
	}, nil
}

// Save the new remote node
func (n *Node) registerRemote(conn net.Conn) {
	node := &remoteNode{
		Data: nil,
		Conn: conn,
	}
	n.remoteNodes[conn.RemoteAddr().String()] = node
	defer conn.Close()
	defer delete(n.remoteNodes, conn.RemoteAddr().String())

	// Start reading
	buf := bufio.NewReader(conn)

	for {
		bytes, err := buf.ReadBytes('\n')
		if err != nil {
			log.Printf("%s<-%s close: %v", n.localAddr, conn.RemoteAddr(), err)
			break
		}
		var msg Message
		if err := msg.Decode(bytes); err != nil {
			log.Printf("%s<-%s msg: %v", n.localAddr, conn.RemoteAddr(), err)
			continue
		}
		n.OutChan <- &msg
	}
}

func (n *Node) Broadcast(msg *Message) error {
	bytes, err := msg.Encode()
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')

	for addr, node := range n.remoteNodes {
		if _, err := node.Conn.Write(bytes); err != nil {
			log.Printf("%s->%s error: %v", n.localAddr, addr, err)
			continue
		}
	}

	return nil
}

func (n *Node) Start() error {
	n.running = true

	for _, addr := range n.remoteAddrs {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("%s->%s error: %v", n.localAddr, addr, err)
			continue
		}
		go n.registerRemote(conn)
	}

	var err error
	if n.ln, err = net.Listen("tcp", n.localAddr); err != nil {
		return err
	}

	go func() {
		defer n.ln.Close()
		for n.running {
			conn, err := n.ln.Accept()
			if err != nil {
				log.Printf("%s error: %v", n.localAddr, err)
				continue
			}
			n.registerRemote(conn)
		}
	}()

	return nil
}

func (n *Node) Stop() {
	n.running = false
	n.ln.Close()
	for _, node := range n.remoteNodes {
		node.Conn.Close()
		delete(n.remoteNodes, node.Conn.RemoteAddr().String())
	}
}
