package types

import (
	"fmt"
	"io"
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
)

// Block is used to build the blockchain
type Block struct {
	number *big.Int
	data   string
}

// "external" block encoding. used for eth protocol, etc.
type extblock struct {
	Number *big.Int
	Data   string
}

// NewBlock create a new bock
func NewBlock(number *big.Int, data string) *Block {
	return &Block{
		number: number,
		data:   data,
	}
}

// RlpHash TODO
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

// EncodeRLP TODO
func (b *Block) EncodeRLP(w io.Writer) error {
	ext := extblock{
		Number: b.number,
		Data:   b.data,
	}
	propToBytes, err := rlp.EncodeToBytes(ext)
	if err != nil {
		return err
	}
	return rlp.Encode(w, ibft.EncodedProposal{
		Type: TypeBlock,
		Prop: propToBytes,
	})
}

// DecodeRLP implements rlp.Decoder, and load the consensus fields from a RLP stream.
func (b *Block) DecodeRLP(s *rlp.Stream) error {
	var eb extblock

	if err := s.Decode(&eb); err != nil {
		return err
	}
	b.number, b.data = eb.Number, eb.Data

	return nil
}
