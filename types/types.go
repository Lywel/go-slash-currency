package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	TypeBlock = iota
)

// RlpHash TODO
func RlpHash(x interface{}) []byte {
	var h common.Hash
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h.Bytes()
}
