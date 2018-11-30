package endpoint

import (
	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-ibft/core"
	"encoding/json"
	"github.com/coryb/gotee"
	"log"
	"net/http"
	"reflect"
)

const logFile = "slash-currency.logs"

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
}

// New returns a new endpoint
func New() *Endpoint {
	ep := &Endpoint{
		broadcast:        make(chan interface{}),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		clients:          make(map[*Client]bool),
		networkMapGetter: nil,
	}

	ep.tee, _ = gotee.NewTee(logFile)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(ep, w, r)
	})

	http.HandleFunc("/hello", ep.helloHandler)
	http.HandleFunc("/logs", ep.logsHandler)

	return ep
}

func (ep *Endpoint) logsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./logs")
}

func (ep *Endpoint) helloHandler(w http.ResponseWriter, r *http.Request) {
	res := json.NewEncoder(w)
	res.Encode("Hello world")
}

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
		log.Fatal("ListenAndServe: ", err)
	}
}

func (ep *Endpoint) handleMsg(msg *message, cli *Client) {
	log.Printf("Received client req: %s", msg.Type)
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

// EventProxy returns a directional channel proxy that forwards core.Event.
// Events are not modified and forwarded as is, this way:
//	in, pout <- pin, out
func (ep *Endpoint) EventProxy() func(in, out chan core.Event) (pin, pout chan core.Event) {
	return func(in, out chan core.Event) (pin, pout chan core.Event) {
		pin = make(chan core.Event, 256)
		pout = make(chan core.Event, 256)
		go func() {
			for i := range pin {
				ep.broadcast <- message{
					Type:     "ibftEventIn",
					Data:     i,
					DataType: reflect.TypeOf(i).String(),
				}
				in <- i
			}
			close(in)
		}()
		go func() {
			for o := range out {
				ep.broadcast <- message{
					Type:     "ibftEventOut",
					Data:     o,
					DataType: reflect.TypeOf(o).String(),
				}
				pout <- o
			}
			close(pout)
		}()
		return
	}
}
