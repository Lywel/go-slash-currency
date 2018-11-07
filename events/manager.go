package events

import (
	"bitbucket.org/ventureslash/go-gossipnet"
	"bitbucket.org/ventureslash/go-ibft/core"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

// Manager handles data from the network
type Manager struct {
	node      *gossipnet.Node
	eventsOut chan core.Event
	eventsIn  chan core.Event
}

type networkMessage struct {
	Type uint
	Data []byte
}

const (
	messageEvent uint = iota
	requestEvent
	backlogEvent
	joinEvent
	stateEvent
)

// New returns a new network.Manager
func New(node *gossipnet.Node, eventsIn, eventsOut chan core.Event) Manager {
	return Manager{
		node:      node,
		eventsIn:  eventsIn,
		eventsOut: eventsOut,
	}
}

// Start Broadcast core address and starts to listen on node.EventChan()
func (mngr Manager) Start() {
	/*
		addrBytes := addr.GetBytes()
		joinBytes, err := rlp.EncodeToBytes(networkMessage{
			Type: joinEvent,
			Data: addrBytes[:],
		})
		if err != nil {
			log.Print("encode error: ", err)
		}
	*/

	// Dispatch network events to IBFT
	go func() {
		for event := range mngr.node.EventChan() {
			switch ev := event.(type) {
			case gossipnet.ConnOpenEvent:
				log.Print("ConnOpenEvent")
				// TODO: dont gossip to everyone, just the new connection
				mngr.node.Gossip([]byte("Slt c mon address :)"))
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
					mngr.eventsIn <- core.MessageEvent{
						Payload: msg.Data,
					}
				case joinEvent:
					log.Print(" -JoinEvent")
					evt := core.JoinEvent{}
					evt.Address.FromBytes(msg.Data)
					if err != nil {
						log.Print(err)
						continue
					}
					mngr.eventsIn <- evt
				case stateEvent:
					log.Print(" -StateEvent")

				}
			case gossipnet.ListenEvent:
				log.Print("ListenEvent")
			case gossipnet.CloseEvent:
				log.Print("CloseEvent")
				close(mngr.eventsIn)
				break
			}
		}
	}()

	// Dispatch IBFT events to the network
	go func() {
		for event := range mngr.eventsOut {
			switch ev := event.(type) {
			case core.MessageEvent:
				mngr.broadcast(ev.Payload)
			}
		}
	}()
}

// Broadcast implements network.Manager.Broadcast. It will tag the payload
// forward it to the network node
func (mngr Manager) broadcast(payload []byte) (err error) {
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
