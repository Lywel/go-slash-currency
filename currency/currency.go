package currency

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-ibft/core"
	"bitbucket.org/ventureslash/go-ibft/crypto"
	"bitbucket.org/ventureslash/go-slash-currency/endpoint"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/logger"
)

const (
	blockInterval    = 5 * time.Second
	blockTimeoutTime = 8 * time.Second
)

var (
	verbose = flag.Bool("verbose-currency", false, "print currency info level logs")

	errInvalidProposal         = errors.New("invalid proposal")
	errInvalidBlock            = errors.New("invalid block hash")
	errUnauthorizedTransaction = errors.New("this transaction is not authorized")
)

type transaction struct {
	From      ibft.Address
	To        ibft.Address
	Amount    *big.Int
	Signature []byte
}

// Currency initializes currency logic
type Currency struct {
	blockchain    []*types.Block
	transactions  []*types.Transaction
	backend       *backend.Backend
	valSet        *ibft.ValidatorSet
	txEvents      chan core.CustomEvent
	endpoint      *endpoint.Endpoint
	mineTimer     *time.Timer
	blockTimeout  *time.Timer
	logger        *logger.Logger
	coreRunning   bool
	waitForValSet bool
	currentSigner uint64
}

// New creates a new currency manager
func New(blockchain []*types.Block, transactions []*types.Transaction, config *backend.Config, privateKey *ecdsa.PrivateKey) *Currency {
	currency := &Currency{
		blockchain:   blockchain,
		transactions: transactions,
		txEvents:     make(chan core.CustomEvent),
		endpoint:     endpoint.New(),
		logger:       logger.Init("Currency", *verbose, false, ioutil.Discard),
	}

	currency.backend = backend.New(config, privateKey, currency, currency.endpoint.EventProxy(), currency.txEvents)
	currency.endpoint.Currency = currency
	currency.endpoint.Backend = currency.backend
	currency.endpoint.SetNetworkMapGetter(currency.backend.Network)
	currency.currentSigner = 0
	currency.coreRunning = false
	currency.waitForValSet = false

	return currency
}

//SyncAndStart synchronize state before startig the currency
func (c *Currency) SyncAndStart(remote string) {
	c.logger.Info("remote addr for sync: ", remote)
	if remote == "" {
		c.Start(true)
	} else {

		r, err := http.Get("http://" + remote + "/state")
		if err != nil {
			c.logger.Warningf("/state request failed: %v", err)
			c.Start(true)
			return
		}

		var state types.State
		err = json.NewDecoder(r.Body).Decode(&state)
		if err != nil {
			c.logger.Warningf("/state invalid response: %v", err)
			c.Start(true)
			return
		}

		c.handleState(state.Blockchain, state.Transactions)
		c.currentSigner = c.blockchain[len(c.blockchain)-1].Number().Uint64()
		c.Start(false)
	}
}

// Start makes the currency manager run
func (c *Currency) Start(isFirstNode bool) {
	c.backend.Start()

	defer c.backend.Stop()
	go c.endpoint.Start(":" + os.Getenv("EP_PORT"))

	if isFirstNode {
		c.createGenesisBlock()
		c.valSet = ibft.NewSet([]ibft.Address{c.backend.Address()})
		c.backend.StartCore(c.valSet, &ibft.View{
			Sequence: big.NewInt(int64(len(c.blockchain) - 1)),
			Round:    ibft.Big0,
		})
	} else {
		c.waitForValSet = true
	}
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

// Commit is called by an IBFT algorithm when a Proposal is accepted
func (c *Currency) Commit(proposal ibft.Proposal) error {
	block, ok := proposal.(*types.Block)
	if !ok {
		return errInvalidProposal
	}
	c.blockchain = append(c.blockchain, block)
	c.transactions = types.TxDifference(c.transactions, block.Transactions)
	if c.blockTimeout != nil {
		c.blockTimeout.Stop()
	}
	c.blockTimeout = time.AfterFunc(blockTimeoutTime, c.handleTimeout)
	c.currentSigner = proposal.Number().Uint64()

	if c.isProposer() {
		c.setTimer()
	}
	return nil
}

func (c *Currency) createGenesisBlock() {
	genesisBlock := types.NewBlock(&types.Header{
		Number:     big.NewInt(0),
		ParentHash: []byte{},
		Time:       big.NewInt(time.Now().Unix()),
	}, types.Transactions{})

	c.blockchain = []*types.Block{genesisBlock}
	c.logger.Info("Genesis block created")
	c.setTimer()
}

func (c *Currency) submitBlock() {
	lastBlock := c.blockchain[len(c.blockchain)-1]
	block := types.NewBlock(&types.Header{
		Number:     new(big.Int).Add(lastBlock.Header.Number, ibft.Big1),
		ParentHash: lastBlock.Hash(),
		Time:       big.NewInt(time.Now().Unix()),
	}, c.transactions)
	c.logger.Info("Mine and submit block: ", block)
	encodedProposal, err := rlp.EncodeToBytes(block)
	if err != nil {
		log.Print(err)
		return
	}
	requestEvent := core.EncodedRequestEvent{
		Proposal: encodedProposal,
	}
	c.backend.EventsOutChan() <- requestEvent

}

func (c *Currency) handleEvent() {
	for event := range c.txEvents {
		switch event.Type {
		case ibft.TypeJoinEvent:
			c.logger.Info("Handling JoinEvent")
			addr := ibft.Address{}
			addr.FromBytes(event.Msg)
			if c.isProposer() {
				c.backend.EventsOutChan() <- core.ValidatorSetEvent{
					ValSet: c.valSet,
					Dest:   addr,
				}
			}
			c.valSet.AddValidator(addr)
		case ibft.TypeValidatorSetEvent:
			c.logger.Info("Handling ValidatorSetEvent")
			valSetEvent := core.ValidatorSetEvent{}
			err := rlp.DecodeBytes(event.Msg, &valSetEvent)
			if err != nil {
				c.logger.Warning("decode ValidatorSetEvent failed")
				continue
			}
			if c.waitForValSet && valSetEvent.Dest == c.backend.Address() {
				c.handleValidatorSetEvent(valSetEvent)
			}
		case ibft.TypeRemoveValidatorEvent:
			c.logger.Info("Handling RemoveValidatorEvent")
			addr := ibft.Address{}
			addr.FromBytes(event.Msg)
			c.valSet.RemoveValidator(addr)
		case ibft.TypeCustomEvents:
			c.logger.Info("Handling txEvent")
			tx := transaction{}
			err := rlp.DecodeBytes(event.Msg, &tx)
			if err != nil {
				c.logger.Warning("decode transaction failed")
				continue
			}
			if err = verifyTransaction(tx); err != nil {
				c.logger.Warning(err)
				continue
			}
			c.logger.Info("Tx verified and added to TxList ", "tx ", tx)
			c.addTransactionToList(tx)
		}

	}
}

func (c *Currency) handleValidatorSetEvent(ev core.ValidatorSetEvent) {
	c.logger.Info("Update valset ", ev.ValSet)
	c.valSet = ev.ValSet
	c.valSet.AddValidator(c.backend.Address())
	c.backend.StartCore(c.valSet, &ibft.View{
		Sequence: big.NewInt(int64(len(c.blockchain) - 1)),
		Round:    ibft.Big0,
	})
	c.waitForValSet = false
	if c.isProposer() {
		lastBlockTimestamp := time.Duration(c.blockchain[len(c.blockchain)-1].Header.Time.Uint64()) * time.Second
		now := time.Duration(time.Now().Unix()) * time.Second
		timeToWait := blockInterval + lastBlockTimestamp - now
		c.logger.Infof("Wait %d before mining", timeToWait)
		c.mineTimer = time.AfterFunc(timeToWait, c.mine)
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
	tx := types.NewTransaction(t.To, t.From, t.Amount)
	c.transactions = append(c.transactions, tx)
}

func (c *Currency) GetState() types.State {
	return types.State{
		Blockchain:   c.blockchain,
		Transactions: c.transactions,
	}
}
