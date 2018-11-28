package types

import (
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
)

type Checkpoint struct {
	BlockNumber       *big.Int
	Balances          map[ibft.Address]*big.Int
	CurrentValidators *ibft.ValidatorSet
}
