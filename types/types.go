package types

import (
	"bitbucket.org/ventureslash/go-slash-currency/blockchain"
)

const (
	// TypeBlock designate a proposal of type block
	TypeBlock = iota
)

// State describes the current currency state
type State struct {
	Blockchain   blockchain.BlockChain `json:"blockchain"`
	Transactions []*Transaction        `json:"transactions"`
}
