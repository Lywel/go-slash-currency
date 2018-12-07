package types

import (
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
)

// Transaction represents a transaction sent over the network
type Transaction struct {
	From      ibft.Address `json:"from"`
	To        ibft.Address `json:"to"`
	Amount    *big.Int     `json:"amount"`
	Signature []byte       `json:"signature"`
}

// NewTransaction initializes a transaction
func NewTransaction(from ibft.Address, to ibft.Address, amount *big.Int) *Transaction {
	return &Transaction{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Hash compute the hash of a transaction
func (s *Transaction) Hash() ibft.Hash {
	return ibft.RlpHash(s)
}

// TxDifference returns a new set which is the difference between a and b.
func TxDifference(a, b Transactions) Transactions {
	keep := make(Transactions, 0, len(a))

	remove := make(map[ibft.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}
