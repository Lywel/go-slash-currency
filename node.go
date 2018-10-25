package gossipnet

import (
	"bufio"
	"crypto/sha256"
	"github.com/hashicorp/golang-lru"
	"io"
	"log"
	"net"
)

const (
	inmemoryPeers    = 40
	inmemoryMessages = 1024
)

// Node is the local Node
type Node struct {
	localAddr      string
	remoteAddrs    []string
	ln             net.Listener
	running        bool
	remoteNodes    map[net.Addr]net.Conn
	readC          chan []byte
	recentMessages *lru.ARCCache // the cache of peer's messages
	knownMessages  *lru.ARCCache // the cache of self messages
}

// New Creates a Network Gossiping Node
func New(localAddr string, remoteAddrs []string) *Node {
	recentMessages, _ := lru.NewARC(inmemoryPeers)
	knownMessages, _ := lru.NewARC(inmemoryMessages)

	return &Node{
		localAddr:      localAddr,
		remoteAddrs:    remoteAddrs,
		running:        false,
		remoteNodes:    make(map[net.Addr]net.Conn),
		readC:          make(chan []byte, 128),
		recentMessages: recentMessages,
		knownMessages:  knownMessages,
	}
}

// Save the new remote node
func (n *Node) registerRemote(conn net.Conn) {
	n.remoteNodes[conn.RemoteAddr()] = conn
	defer conn.Close()
	defer delete(n.remoteNodes, conn.RemoteAddr())

	// Start reading
	buf := bufio.NewReader(conn)

	for {
		payload, err := buf.ReadBytes('\n')
		switch err {
		case nil:
			n.handleData(conn.RemoteAddr().String(), payload[:len(payload)-1])
			continue
		case io.EOF:
		default:
			log.Print(err)
		}
		break
	}
}

func (n *Node) handleData(addr string, payload []byte) {
	hash := sha256.Sum256(payload)
	n.cacheEventFor(addr, hash)

	if _, alreadyKnew := n.knownMessages.Get(hash); alreadyKnew {
		return
	}
	n.knownMessages.Add(hash, true)

	n.Gossip(payload)

	// protection against blocking channel
	select {
	case n.readC <- payload:
	default:
	}
}

func (n *Node) cacheEventFor(addr string, hash [32]byte) (alreadyKnew bool) {
	cached, hasCache := n.recentMessages.Get(addr)
	var recentMsgs *lru.ARCCache
	if hasCache {
		recentMsgs, _ = cached.(*lru.ARCCache)
		_, alreadyKnew = recentMsgs.Get(hash)
	} else {
		recentMsgs, _ = lru.NewARC(inmemoryMessages)
	}
	recentMsgs.Add(hash, true)
	n.recentMessages.Add(addr, recentMsgs)
	return
}

// EventChan returns a readonly chanel for data events
func (n *Node) EventChan() <-chan []byte {
	return n.readC
}

// Gossip sends a Message to all peers passing selection (except self)
func (n *Node) Gossip(payload []byte) {
	hash := sha256.Sum256(payload)
	payload = append(payload, '\n')

	for addr, conn := range n.remoteNodes {
		alreadyKnew := n.cacheEventFor(addr.String(), hash)
		if !alreadyKnew {
			conn.Write(payload)
		}
	}
}

// Broadcast sends a Message to all peers passing selection (including self)
func (n *Node) Broadcast(payload []byte) {
	n.Gossip(payload)
	n.handleData(n.localAddr, payload)
}

// Start starts the node (client / server)
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
			go n.registerRemote(conn)
		}
	}()

	return nil
}

// Stop closes all connection and stops listening
func (n *Node) Stop() {
	n.running = false
	n.ln.Close()
	for _, conn := range n.remoteNodes {
		conn.Close()
	}
}
