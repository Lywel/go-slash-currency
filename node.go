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
func New(localAddr string, remoteAddrs []string) (*Node, error) {
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
	}, nil
}

func (n *Node) recv(addr string, payload []byte) {
	hash := sha256.Sum256(payload)
	// Mark peer's message
	cache, hasCache := n.recentMessages.Get(addr)
	var recentMsgs *lru.ARCCache
	if hasCache {
		recentMsgs, _ = cache.(*lru.ARCCache)
	} else {
		recentMsgs, _ = lru.NewARC(inmemoryMessages)
		n.recentMessages.Add(addr, recentMsgs)
	}
	recentMsgs.Add(hash, true)

	// Mark self known message
	if _, alreadyKnew := n.knownMessages.Get(hash); alreadyKnew {
		return
	}

	n.knownMessages.Add(hash, true)
	n.Gossip(nil, payload)

	// protection against blocking channel
	select {
	case n.readC <- payload:
	default:
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
			n.recv(conn.RemoteAddr().String(), payload[:len(payload)-1])
			continue
		case io.EOF:
		default:
			log.Print(err)
		}
		break
	}
}

// nodeManager is an interface which filters used connections
type nodeManager interface {
	IsInteressted(net.Addr) bool
}

// DefaultNodeManager is a node manager that talks to everyone
type defaultNodeManager struct{}

func (n defaultNodeManager) IsInteressted(net.Addr) bool {
	return true
}

// ReadC returns a readonly chanel on which received messages will be funneled
func (n *Node) ReadC() <-chan []byte {
	return n.readC
}

// Gossip sends a Message to all peers passing selection (except self)
func (n *Node) Gossip(manager nodeManager, payload []byte) {
	if manager == nil {
		manager = defaultNodeManager{}
	}
	hash := sha256.Sum256(payload)
	payload = append(payload, '\n')

	for addr, conn := range n.remoteNodes {
		if manager.IsInteressted(addr) {
			cached, hasCache := n.recentMessages.Get(addr.String())
			var recentMsgs *lru.ARCCache
			if hasCache {
				recentMsgs, _ = cached.(*lru.ARCCache)
				if _, hasMsg := recentMsgs.Get(hash); hasMsg {
					// This peer had this event, skip it
					continue
				}
			} else {
				// Create cache for this peer
				recentMsgs, _ = lru.NewARC(inmemoryMessages)
			}

			recentMsgs.Add(hash, true)
			n.recentMessages.Add(addr.String(), recentMsgs)

			if _, err := conn.Write(payload); err != nil {
				log.Printf("%s->%s error: %v", n.localAddr, addr, err)
				continue
			}
		}
	}
}

// Broadcast sends a Message to all peers passing selection (including self)
func (n *Node) Broadcast(manager nodeManager, payload []byte) {
	n.Gossip(manager, payload)
	n.recv(n.localAddr, payload)
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
