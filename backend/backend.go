package backend

import (
	"crypto/ecdsa"
	"bitbucket.org/ventureslash/go-gossipnet"
	"bitbucket.org/ventureslash/go-ibft/consensus"
	"bitbucket.org/ventureslash/go-ibft/consensus/backend/crypto"
	"bitbucket.org/ventureslash/go-ibft/consensus/core"
)

// Backend initializes the core, holds the keys and currenncy logic
type backend struct {
	privateKey  *ecdsa.PrivateKey
	address     consensus.Address
	network     *gossipnet.Node
	core        consensus.Engine
	coreRunning bool
	valSet      *consensus.ValidatorSet
}

// Config is the backend configuration struct
type Config struct {
	LocalAddr   string
	RemoteAddrs []string
}

// New returns a new Backend
func New(config *Config, privateKey *ecdsa.PrivateKey) consensus.Backend {
	network := gossipnet.New(config.LocalAddr, config.RemoteAddrs)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	backend := &backend{
		privateKey: privateKey,
		address:    address,
		network:    network,
		valSet:     consensus.NewSet([]consensus.Address{address}),
	}

	backend.core = core.New(backend)
	return backend
}

func (b *backend) Address() consensus.Address {
	return b.address
}

func (b *backend) Network() *gossipnet.Node {
	return b.network
}

func (b *backend) AddValidator(addr consensus.Address) bool {
	return b.valSet.AddValidator(addr)
}

// Start implements Engine.Start
func (b *backend) Start() {
	if b.coreRunning {
		return
	}
	b.network.Start()
	b.core.Start()
	b.coreRunning = true
}

// Stop implements Engine.Stop
func (b *backend) Stop() {
	if !b.coreRunning {
		return
	}
	b.network.Stop()
	b.core.Stop()
	b.coreRunning = false
}

// Sign implements Backend.Sign
func (b *backend) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, b.privateKey)
}
