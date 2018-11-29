package currency

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"log"
	"math/big"
	"os"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-ibft/crypto"
	"bitbucket.org/ventureslash/go-slash-currency/endpoint"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	errInvalidProposal         = errors.New("invalid proposal")
	errInvalidBlock            = errors.New("invalid block hash")
	errUnauthorizedTransaction = errors.New("this transaction is not authorized")
)

type transaction struct {
	From      ibft.Address
	To        ibft.Address
	Amount    uint64
	Signature []byte
}

// Currency initializes currency logic
type Currency struct {
	blockchain   []*types.Block
	transactions []*types.Transaction
	backend      *backend.Backend
	txEvents     chan []byte
	endpoint     *endpoint.Endpoint
}

// New creates a new currency manager
func New(blockchain []*types.Block, transactions []*types.Transaction, config *backend.Config, privateKey *ecdsa.PrivateKey) *Currency {
	currency := &Currency{
		blockchain:   blockchain,
		transactions: transactions,
		txEvents:     make(chan []byte),
		endpoint:     endpoint.New(),
	}

	currency.backend = backend.New(config, privateKey, currency, currency.endpoint.EventProxy(), currency.txEvents)
	currency.endpoint.SetNetworkMapGetter(currency.backend.Network)

	log.Print("configured to run on port: " + os.Getenv("VAL_PORT"))
	return currency
}

// Start makes the currency manager run
func (c *Currency) Start() {
	c.backend.Start()
	defer c.backend.Stop()

	go c.endpoint.Start(":" + os.Getenv("EP_PORT"))
	c.handleEvent()
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

func (c *Currency) CreateBlock() {}

func (c *Currency) handleEvent() {
	for event := range c.txEvents {

		tx := transaction{}
		err := rlp.DecodeBytes(event, &tx)
		if err != nil {
			log.Print("decode transaction failed")
			continue
		}
		if err = verifyTransaction(tx); err != nil {
			log.Print(err)
			continue
		}
		c.addTransactionToList(tx)
	}
}

func verifyTransaction(t transaction) error {
	txNoSig := transaction{
		From:      t.From,
		To:        t.To,
		Amount:    t.Amount,
		Signature: []byte{},
	}

	txNoSigRlp, err := rlp.EncodeToBytes(txNoSig)
	if err != nil {
		return err
	}

	addressFrom, err := crypto.GetSignatureAddress(txNoSigRlp, t.Signature)
	if err != nil {
		return err
	}
	if addressFrom != t.From {
		return errUnauthorizedTransaction
	}
	return nil
}

func (c *Currency) addTransactionToList(t transaction) {
	tx := types.NewTransaction(t.To, t.From, big.NewInt(int64(t.Amount)))
	c.transactions = append(c.transactions, tx)
}
