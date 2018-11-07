package network

import (
	"bitbucket.org/ventureslash/go-gossipnet"
	"bitbucket.org/ventureslash/go-ibft/consensus"
	"bitbucket.org/ventureslash/go-ibft/events"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

// Manager handles data from the network
type Manager struct {
	node   *gossipnet.Node
	events events.Handler
}

type networkMessage struct {
	Type uint
	Data []byte
}

// MessageEvent is emmitted during the IBFT consensus algo
type MessageEvent struct {
	Payload []byte
}

// JoinEvent is emmitted when a peer joins the network
type JoinEvent struct {
	Address consensus.Address
}

const (
	messageEvent uint = iota
	requestEvent
	backlogEvent
	joinEvent
	stateEvent
)

// New returns a new network.Manager
func New(node *gossipnet.Node, events events.Handler) Manager {
	return Manager{
		node:   node,
		events: events,
	}
}

// Start Broadcast core address and starts to listen on node.EventChan()
func (mngr Manager) Start(addr consensus.Address) {
	addrBytes := addr.GetBytes()
	joinBytes, err := rlp.EncodeToBytes(networkMessage{
		Type: joinEvent,
		Data: addrBytes[:],
	})
	if err != nil {
		log.Print("encode error: ", err)
	}

	go func() {
		for event := range mngr.node.EventChan() {
			switch ev := event.(type) {
			case gossipnet.ConnOpenEvent:
				log.Print("ConnOpenEvent")
				// TODO: dont gossip to everyone, just the new connection
				mngr.node.Gossip(joinBytes)
			case gossipnet.ConnCloseEvent:
				log.Print("ConnCloseEvent")
			case gossipnet.DataEvent:
				log.Print("DataEvent")
				var msg networkMessage
				err := rlp.DecodeBytes(ev.Data, &msg)
				if err != nil {
					log.Print("Error parsing msg:", string(ev.Data))
					continue
				}
				switch msg.Type {
				case messageEvent:
					log.Print(" -MsgEvent")
					mngr.events.Push(MessageEvent{
						Payload: msg.Data,
					})
				case joinEvent:
					log.Print(" -JoinEvent")
					evt := JoinEvent{}
					evt.Address.FromBytes(msg.Data)
					if err != nil {
						log.Print(err)
						continue
					}
					mngr.events.Push(evt)
				case stateEvent:
					log.Print(" -StateEvent")

				}
			case gossipnet.ListenEvent:
				log.Print("ListenEvent")
			case gossipnet.CloseEvent:
				log.Print("CloseEvent")
				mngr.events.Close()
				break
			}
		}
	}()
}

// Broadcast implements network.Manager.Broadcast. It will tag the payload
// forward it to the network node
func (mngr Manager) Broadcast(payload []byte) (err error) {
	data, err := rlp.EncodeToBytes(networkMessage{
		Type: messageEvent,
		Data: payload,
	})
	if err != nil {
		return
	}
	mngr.node.Broadcast(data)
	return
}
