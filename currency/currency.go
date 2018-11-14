package currency

import (
	"errors"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Currency initializes currency logic
type Currency struct {
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
func (c *Currency) Verify(proposal ibft.Proposal) error { return nil }

// Commit is called by an IBFT algorythm when a Proposal is accepted
func (c *Currency) Commit(proposal ibft.Proposal) error { return nil }
