package types

import (
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
)

type Checkpoint struct {
	BlockNumber       *big.Int
	balances          map[ibft.Address]*big.Int
	currentValidators *ibft.ValidatorSet
}
