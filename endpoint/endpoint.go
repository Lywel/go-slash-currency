package endpoint

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"reflect"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-ibft/core"
	"bitbucket.org/ventureslash/go-slash-currency/blockchain"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/coryb/gotee"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/logger"
)

type currency interface {
	DecodeProposal(*ibft.EncodedProposal) (ibft.Proposal, error)
	BlockChain() *blockchain.BlockChain
	PendingTransactions() []*types.Transaction
}

const logFile = "slash-currency.logs"

var verbose = flag.Bool("verbose-endpoint", false, "print endpoint info level logs")

// Endpoint maintains the set of active clients and broadcasts messages to the
// clients.
type Endpoint struct {
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan interface{}
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
	// A function that returns a mapping of connected clients
	networkMapGetter func() map[ibft.Address]string
	// log tee
	tee *gotee.Tee

	Currency currency
	Backend  *backend.Backend
	debug    *logger.Logger
}

// New returns a new endpoint
func New() *Endpoint {
	ep := &Endpoint{
		broadcast:        make(chan interface{}),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		clients:          make(map[*Client]bool),
		networkMapGetter: nil,
		debug:            logger.Init("Endpoint", *verbose, false, ioutil.Discard),
	}

	ep.tee, _ = gotee.NewTee(logFile)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(ep, w, r)
	})

	http.HandleFunc("/hello", ep.helloHandler)
	http.HandleFunc("/logs", ep.logsHandler)
	http.HandleFunc("/state", ep.stateHandler)

	return ep
}

func (ep *Endpoint) logsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, logFile)
}

func (ep *Endpoint) stateHandler(w http.ResponseWriter, r *http.Request) {
	state := struct {
		Blockchain   *blockchain.BlockChain
		Transactions []*types.Transaction
	}{}

	state.Blockchain = ep.Currency.BlockChain()
	state.Transactions = ep.Currency.PendingTransactions()

	err := rlp.Encode(w, state)
	if err != nil {
		ep.debug.Warningf("failed to encode state: %v", err)
	}
}

func (ep *Endpoint) helloHandler(w http.ResponseWriter, r *http.Request) {
	res := json.NewEncoder(w)
	res.Encode("Hello world")
}

// SetNetworkMapGetter sets the networkmap getter
func (ep *Endpoint) SetNetworkMapGetter(networkMapGetter func() map[ibft.Address]string) {
	ep.networkMapGetter = networkMapGetter
}

func (ep *Endpoint) run() {
	for {
		select {
		case client := <-ep.register:
			ep.clients[client] = true
		case client := <-ep.unregister:
			if _, ok := ep.clients[client]; ok {
				delete(ep.clients, client)
				close(client.send)
			}
		case message := <-ep.broadcast:
			for client := range ep.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(ep.clients, client)
				}
			}
		}
	}
}

// Start starts the endpoint
func (ep *Endpoint) Start(addr string) {
	go ep.run()

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		ep.debug.Errorf("ListenAndServe: %v", err)
	}
}

func (ep *Endpoint) handleMsg(msg *message, cli *Client) {
	ep.debug.Infof("Received client req: %s", msg.Type)
	switch msg.Type {
	case "network-state":
		if ep.networkMapGetter == nil {
			cli.send <- message{
				Type: "error",
				Data: "Network getter is not configured on this server: nil",
			}
			return
		}
		network := ep.networkMapGetter()

		netmap := make(map[string]string)

		for k, v := range network {
			netmap[k.String()] = v
		}

		cli.send <- message{
			Type: "network-state",
			Data: netmap,
		}

	}
}

func (ep *Endpoint) publishEvent(e interface{}, eType string) {
	msg := message{
		Type: eType,
		Data: e,
	}

	switch ev := e.(type) {
	case core.MessageEvent:
		return
	case core.EncodedRequestEvent:
		if ep.Currency == nil {
			//ep.debug.Warning("Unable to decode Request event: ep.currency is nil")
			break
		}

		// Proposal: []byte -> EncodedProposal
		var encodedProposal *ibft.EncodedProposal
		err := rlp.DecodeBytes(ev.Proposal, &encodedProposal)
		if err != nil {
			break
		}
		// encodedProposal: EncodedProposal -> Proposal
		proposal, err := ep.Currency.DecodeProposal(encodedProposal)
		if err != nil {
			break
		}
		msg.Data = core.RequestEvent{Proposal: proposal}
		break
	case []byte:
		tx := types.Transaction{}
		err := rlp.DecodeBytes(ev, &tx)
		if err != nil {
			break
		}
		msg.Data = tx
	}

	msg.DataType = reflect.TypeOf(msg.Data).String()
	ep.broadcast <- msg
}

// EventProxy returns a directional channel proxy that forwards core.Event.
// Events are not modified and forwarded as is, this way:
//	in, pout <- pin, out
func (ep *Endpoint) EventProxy() func(in, out chan core.Event, custom chan core.CustomEvent) (pin, pout chan core.Event, pcustom chan core.CustomEvent) {
	return func(in, out chan core.Event, custom chan core.CustomEvent) (pin, pout chan core.Event, pcustom chan core.CustomEvent) {
		pin = make(chan core.Event, 256)
		pout = make(chan core.Event, 256)
		pcustom = make(chan core.CustomEvent, 256)
		go func() {
			for i := range pin {
				ep.publishEvent(i, "ibftEventIn")
				in <- i
			}
			close(in)
		}()
		go func() {
			for o := range out {
				ep.publishEvent(o, "ibftEventOut")
				pout <- o
			}
			close(pout)
		}()
		go func() {
			for c := range pcustom {
				ep.publishEvent(c, "txEvent")
				custom <- c
			}
			close(in)
		}()
		return
	}
}
