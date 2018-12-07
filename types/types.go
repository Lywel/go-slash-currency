package types

import ()

const (
	// TypeBlock designate a proposal of type block
	TypeBlock = iota
)

// State describes the current currency state
type State struct {
	Blockchain   []*Block       `json:"blockchain"`
	Transactions []*Transaction `json:"transactions"`
}

type Receipt struct {
}
type Receipts []*Receipt
