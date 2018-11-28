package currency

import (
	"bytes"
	"errors"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	errInvalidProposal = errors.New("invalid proposal")
	errInvalidBlock    = errors.New("invalid block hash")
)

// Currency initializes currency logic
type Currency struct {
	blockchain   []*types.Block
	transactions []*types.Transaction
}

// DecodeProposal parses a payload and return a Proposal interface
func (c *Currency) DecodeProposal(prop *ibft.EncodedProposal) (ibft.Proposal, error) {
	switch prop.Type {
	case types.TypeBlock:
		var b *types.Block
		err := rlp.DecodeBytes(prop.Prop, &b)
		if err != nil {
			return nil, err
		}
		return b, nil
		/*
			case type.Transactiono:
				return proposal
		*/
	default:
		return nil, errors.New("Unknown proposal type")
	}
}

// Verify returns an error is a proposal should be rejected
func (c *Currency) Verify(proposal ibft.Proposal) error {
	block, ok := proposal.(*types.Block)
	if !ok {
		return errInvalidProposal
	}
	lastBlock := c.blockchain[len(c.blockchain)-1]
	if bytes.Compare(block.Header.ParentHash, lastBlock.Hash()) != 0 {
		return errInvalidBlock
	}
	return nil
}

// Commit is called by an IBFT algorythm when a Proposal is accepted
func (c *Currency) Commit(proposal ibft.Proposal) error {
	block, ok := proposal.(*types.Block)
	if !ok {
		return errInvalidProposal
	}
	c.blockchain = append(c.blockchain, block)
	c.transactions = types.TxDifference(c.transactions, block.Transactions)
	return nil
}
