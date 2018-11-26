package types

import (
	"fmt"
	"io"
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
	"github.com/ethereum/go-ethereum/rlp"
)

type Header struct {
	Number *big.Int
	Time   *big.Int
}

// Block is used to build the blockchain
type Block struct {
	header       *Header
	transactions Transactions
}

// "external" block encoding. used for eth protocol, etc.
type extblock struct {
	Header       *Header
	Transactions Transactions
}

// NewBlock create a new bock
func NewBlock(header *Header, transactions []*Transaction) *Block {
	return &Block{
		header:       header,
		transactions: transactions,
	}
}

// Hash compute the hash of a block
func (b *Block) Hash() []byte {
	return RlpHash(b.header)
}

// Number return the number of a block
func (b *Block) Number() *big.Int {
	return new(big.Int).Set(b.header.Number)
}

func (b *Block) String() string {
	return fmt.Sprintf("number %d, data %s", b.Number())
}

// EncodeRLP TODO
func (b *Block) EncodeRLP(w io.Writer) error {
	ext := extblock{
		Header:       b.header,
		Transactions: b.transactions,
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
	b.header, b.transactions = eb.Header, eb.Transactions

	return nil
}
