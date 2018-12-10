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

// Number return the number of a block
func (b *Block) Number() *big.Int {
	return new(big.Int).Set(b.Header.Number)
}

func (b *Block) String() string {
	return fmt.Sprintf("number %d", b.Number())
}

// EncodeRLP TODO
func (b *Block) EncodeRLP(w io.Writer) error {
	ext := extblock{
		Header:       b.Header,
		Transactions: b.Transactions,
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
	b.Header, b.Transactions = eb.Header, eb.Transactions

	return nil
}
