package endpoint

import (
	"bitbucket.org/ventureslash/go-ibft/core"
	"log"
	"net/http"
	"reflect"
)

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
}

// New returns a new endpoint
func New() *Endpoint {
	ep := &Endpoint{
		broadcast:  make(chan interface{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(ep, w, r)
	})

	return ep
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

// EventProxy returns a directional channel proxy that forwards core.Event.
// Events are not modified and forwarded as is, this way:
//	in, pout <- pin, out
func (ep *Endpoint) EventProxy() func(in, out chan core.Event) (pin, pout chan core.Event) {
	return func(in, out chan core.Event) (pin, pout chan core.Event) {
		pin = make(chan core.Event, 256)
		pout = make(chan core.Event, 256)
		go func() {
			for i := range pin {
				ep.broadcast <- ibftEvent{
					Direction: inboundDir,
					Type:      reflect.TypeOf(i).String(),
					Data:      nil,
				}
				in <- i
			}
			close(in)
		}()
		go func() {
			for o := range out {
				ep.broadcast <- ibftEvent{
					Direction: outboundDir,
					Type:      reflect.TypeOf(o).String(),
					Data:      nil,
				}
				pout <- o
			}
			close(pout)
		}()
		return
	}
}
