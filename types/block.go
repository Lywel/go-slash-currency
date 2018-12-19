package types

import (
	"fmt"
	"io"
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
	"github.com/ethereum/go-ethereum/rlp"
)

type Header struct {
	Number     *big.Int  `json:"number"`
	ParentHash ibft.Hash `json:"parenthash"`
	Time       *big.Int  `json:"timestamp"`
}

// Block is used to build the blockchain
type Block struct {
	Header       *Header      `json:"header"`
	Transactions Transactions `json:"transactions"`
}

// "external" block encoding. used for eth protocol, etc.
type extblock struct {
	Header       *Header
	Transactions Transactions
}

// NewBlock create a new bock
func NewBlock(header *Header, transactions []*Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: transactions,
	}
}

// Hash compute the hash of a block
func (b *Block) Hash() ibft.Hash {
	return ibft.RlpHash(b)
}

// ParentHash returns the parentHash stored in the block header
func (b *Block) ParentHash() ibft.Hash {
	return b.Header.ParentHash
}

// Number return the number of a block
func (b *Block) Number() *big.Int {
	return new(big.Int).Set(b.Header.Number)
}

func (b *Block) String() string {
	return fmt.Sprintf("number %d", b.Number())
}

// ExportAsRLPEncodedProposal exports a block as an rlp encoeded proposal
func (b *Block) ExportAsRLPEncodedProposal() ([]byte, error) {
	propToBytes, err := rlp.EncodeToBytes(b)
	if err != nil || propToBytes == nil {
		return nil, err
	}
	return rlp.EncodeToBytes(ibft.EncodedProposal{
		Type: TypeBlock,
		Prop: propToBytes,
	})
}

// EncodeRLP export a block to rlp stream
func (b *Block) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, extblock{
		Header:       b.Header,
		Transactions: b.Transactions,
	})
}

// DecodeRLP decodes a block from an rlp stream
func (b *Block) DecodeRLP(s *rlp.Stream) error {
	var ext extblock
	if err := s.Decode(&ext); err != nil {
		return err
	}
	b.Header, b.Transactions = ext.Header, ext.Transactions
	return nil
}
