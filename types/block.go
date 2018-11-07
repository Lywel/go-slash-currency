package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
)

type Block struct {
	number *big.Int
	data   string
}

// NewBlock create a new bock
func NewBlock(number *big.Int, data string) *Block {
	return &Block{
		number: number,
		data:   data,
	}
}

func RlpHash(x interface{}) []byte {
	var h common.Hash
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h.Bytes()
}

// Hash compute the hash of a block
func (b *Block) Hash() []byte {
	return RlpHash(b)
}

// Number return the number of a block
func (b *Block) Number() *big.Int {
	return b.number
}

func (b *Block) String() string {
	return fmt.Sprintf("number %d, data %s", b.number, b.data)
}
