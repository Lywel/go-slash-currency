package backend

import (
	"bitbucket.org/ventureslash/go-gossipnet"
	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-ibft/core"
	"bitbucket.org/ventureslash/go-ibft/crypto"
	"bitbucket.org/ventureslash/go-slash-currency/events"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"crypto/ecdsa"
	"errors"
)

// Backend initializes the core, holds the keys and currenncy logic
type Backend struct {
	privateKey    *ecdsa.PrivateKey
	address       ibft.Address
	network       *gossipnet.Node
	core          ibft.Engine
	coreRunning   bool
	ibftEventsIn  chan core.Event
	ibftEventsOut chan core.Event
	manager       events.Manager
}

// Config is the backend configuration struct
type Config struct {
	LocalAddr   string
	RemoteAddrs []string
}

// New returns a new Backend
func New(config *Config, privateKey *ecdsa.PrivateKey) *Backend {
	network := gossipnet.New(config.LocalAddr, config.RemoteAddrs)
	in := make(chan core.Event, 256)
	out := make(chan core.Event, 256)

	backend := &Backend{
		privateKey:    privateKey,
		address:       crypto.PubkeyToAddress(privateKey.PublicKey),
		network:       network,
		ibftEventsIn:  in,
		ibftEventsOut: out,
		manager:       events.New(network, in, out),
	}

	backend.core = core.New(backend)
	return backend
}

// PrivateKey returns the private key
func (b *Backend) PrivateKey() *ecdsa.PrivateKey {
	return b.privateKey
}

// Start implements Engine.Start
func (b *Backend) Start() {
	if b.coreRunning {
		return
	}
	b.manager.Start(b.address)
	b.network.Start()
	b.core.Start()
	b.coreRunning = true
}

// Stop implements Engine.Stop
func (b *Backend) Stop() {
	if !b.coreRunning {
		return
	}
	b.network.Stop()
	b.core.Stop()
	b.coreRunning = false
}

// EventsInChan returns a channel receiving network events
func (b *Backend) EventsInChan() chan core.Event {
	return b.ibftEventsIn
}

// EventsOutChan returns a channel used to emit events to the network
func (b *Backend) EventsOutChan() chan core.Event {
	return b.ibftEventsOut
}

// DecodeProposal parses a payload and return a Proposal interface
func (b *Backend) DecodeProposal(prop interface{}) (ibft.Proposal, error) {
	switch proposal := prop.(type) {
	case *types.Block:
		return proposal, nil
		/*
			case type.Transactiono:
				return proposal
		*/
	default:
		return nil, errors.New("Unknown proposal type")
	}
}

// Verify returns an error is a proposal should be rejected
func (b *Backend) Verify(proposal ibft.Proposal) error { return nil }

// Commit is called by an IBFT algorythm when a Proposal is accepted
func (b *Backend) Commit(proposal ibft.Proposal) error { return nil }
