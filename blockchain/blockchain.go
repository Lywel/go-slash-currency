package blockchain

import (
	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/rawdb"
	"bitbucket.org/ventureslash/go-slash-currency/state"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/logger"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"io/ioutil"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

var verbose = flag.Bool("verbose-blockchain", false, "print blockchain info level logs")

// BlockChain is the structure managing and storing blocks
type BlockChain struct {
	db           *leveldb.DB
	genesisBlock *types.Block
	currentBlock atomic.Value
	mu           sync.RWMutex // global mutex for locking chain operations
	chainmu      sync.RWMutex // blockchain insertion lock
	state        *state.StateDB
	debug        *logger.Logger
}

// New resturns a new instance of Blockchain
func New(file string) (*BlockChain, error) {
	db, err := rawdb.InitDB(file)
	if err != nil {
		return nil, err
	}

	bc := &BlockChain{
		db:    db,
		state: state.New(),
		debug: logger.Init("BlockChain", *verbose, false, ioutil.Discard),
	}

	bc.genesisBlock = bc.readOrCreateGenesisBlock()

	return bc, nil
}

func (bc *BlockChain) readOrCreateGenesisBlock() *types.Block {
	genesis := bc.GetBlockByNumber(0)
	if genesis == nil {
		bc.debug.Info("No genesis block, creating one.")
		genesis = types.NewBlock(&types.Header{
			Number:     big.NewInt(0),
			ParentHash: ibft.Hash{},
			Time:       big.NewInt(time.Now().Unix()),
		}, types.Transactions{})

		rawdb.WriteBlock(bc.db, genesis)
		rawdb.WriteReceipts(bc.db, genesis.Hash(), genesis.Number().Uint64(), nil)
		rawdb.WriteBlockHash(bc.db, genesis.Hash(), genesis.Number().Uint64())
		rawdb.WriteHeadBlockHash(bc.db, genesis.Hash())
	}
	bc.currentBlock.Store(genesis)

	return genesis
}

// GetBlockByHash retrieves a block from the database by hash
func (bc *BlockChain) GetBlockByHash(hash ibft.Hash) *types.Block {
	bc.debug.Infof("GetBlockByHash (%s)", hash)
	number := rawdb.ReadBlockNumber(bc.db, hash)
	if number == nil {
		bc.debug.Warningf("Unable to find block (%s) number", hash)
		return nil
	}
	return bc.GetBlock(hash, *number)
}

// GetBlockByNumber retrieves a block from the database by number
func (bc *BlockChain) GetBlockByNumber(number uint64) *types.Block {
	bc.debug.Infof("GetBlockByNumber (%d)", number)
	hash := rawdb.ReadBlockHash(bc.db, number)
	if hash == (ibft.Hash{}) {
		bc.debug.Warningf("Unable to find block (%d) hash ", number)
		return nil
	}
	return bc.GetBlock(hash, number)
}

// GetBlock retrieves a block from the database by hash and number
func (bc *BlockChain) GetBlock(hash ibft.Hash, number uint64) *types.Block {
	bc.debug.Infof("GetBlock (%d, %s)", number, hash)
	block := rawdb.ReadBlock(bc.db, hash, number)
	if block == nil {
		bc.debug.Warningf("Unable to find block (%d, %s)", number, hash)
		return nil
	}
	return block
}

// WriteBlock writes the block to the database
func (bc *BlockChain) WriteBlock(block *types.Block, receipts []*types.Receipt) error {
	bc.debug.Infof("WriteBlock (%d, %s)", block.Number().Uint64(), block.Hash())
	// Make sure no inconsistent state is leaked during insertion
	bc.mu.Lock()
	defer bc.mu.Unlock()

	rawdb.WriteBlock(bc.db, block)
	// Write the metadata for transaction/receipt lookups and preimages
	rawdb.WriteReceipts(bc.db, block.Hash(), block.Number().Uint64(), receipts)
	// TODO: rawdb.WriteTxLookupEntries(batch, block) # USING A batch !!

	bc.insert(block)
	return nil
}

// insert injects a new head block into the current block chain. This method
// assumes that the block is indeed a true head. It will update currenctHead
// Note, this function assumes that the `mu` mutex is held!
func (bc *BlockChain) insert(block *types.Block) {
	// Add the block to the canonical chain number scheme and mark as the head
	rawdb.WriteBlockHash(bc.db, block.Hash(), block.Number().Uint64())
	rawdb.WriteHeadBlockHash(bc.db, block.Hash())

	bc.currentBlock.Store(block)
}

// CurrentBlock returns the head of the blockchain
func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}

// HasBlock checks if a block is fully present in the database or not.
func (bc *BlockChain) HasBlock(hash ibft.Hash, number uint64) bool {
	return rawdb.HasBlock(bc.db, hash, number)
}

// Export writes the active chain to the given writer.
func (bc *BlockChain) Export(w io.Writer) error {
	return bc.ExportN(w, uint64(0), bc.CurrentBlock().Number().Uint64())
}

// ExportN writes a subset of the active chain to the given writer.
func (bc *BlockChain) ExportN(w io.Writer, first uint64, last uint64) error {
	if first > last {
		return fmt.Errorf("export failed: first (%d) is greater than last (%d)", first, last)
	}
	bc.debug.Infof("Exporting blocks from %d to %d.", first, last)

	/* start, reported := time.Now(), time.Now() */
	for nr := first; nr <= last; nr++ {
		block := bc.GetBlockByNumber(nr)
		if block == nil {
			bc.debug.Errorf("export failed on #%d: not found", nr)
			return fmt.Errorf("export failed on #%d: not found", nr)
		}
		if err := rlp.Encode(w, block); err != nil {
			bc.debug.Errorf("encode block failed on #%d: %v", nr, err)
			return err
		}
		/* if time.Since(reported) >= statsReportLimit { */
		/* 	log.Info("Exporting blocks", "exported", block.Number().Uint64()-first, "elapsed", common.PrettyDuration(time.Since(start))) */
		/* 	reported = time.Now() */
		/* } */
	}

	return nil
}

// EncodeRLP implements encodeRLPer
func (bc *BlockChain) EncodeRLP(w io.Writer) error {
	return bc.Export(w)
}

// InsertChain attempts to complete an already existing header chain with
// transaction.
func (bc *BlockChain) InsertChain(blockChain []*types.Block) error {
	// Sanity check that we have something meaningful to import
	if len(blockChain) == 0 {
		return nil
	}

	// Check if the first block is a child of the current head
	if blockChain[0].ParentHash() != bc.CurrentBlock().Hash() || blockChain[0].Number().Uint64() != bc.CurrentBlock().Number().Uint64() {
		return fmt.Errorf("non continous insert: first block hash is not a child of current head")
	}

	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(blockChain); i++ {
		if blockChain[i].Number().Uint64() != blockChain[i-1].Number().Uint64()+1 || blockChain[i].ParentHash() != blockChain[i-1].Hash() {
			/* log.Error("Non contiguous receipt insert", "number", blockChain[i].Number(), "hash", blockChain[i].Hash(), "parent", blockChain[i].ParentHash(), */
			/* 	"prevnumber", blockChain[i-1].Number(), "prevhash", blockChain[i-1].Hash()) */
			return fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, blockChain[i-1].Number().Uint64(),
				blockChain[i-1].Hash().Bytes()[:4], i, blockChain[i].Number().Uint64(), blockChain[i].Hash().Bytes()[:4], blockChain[i].ParentHash().Bytes()[:4])
		}
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()
	for _, block := range blockChain {
		receipts, err := bc.state.ProcessBlock(block)
		if err != nil {
			return err
		}
		// Write all the data out into the database
		bc.WriteBlock(block, receipts)
	}
	return nil
}

// State returns the current HEAD state
func (bc *BlockChain) State() *state.StateDB {
	return bc.state
}

// ResetWithGenesis purges the entire blockchain, restoring it to the
// specified genesis state.
func (bc *BlockChain) ResetWithGenesis(genesis *types.Block) error {
	// Dump the entire block chain and purge the caches
	if err := bc.SetHead(0); err != nil {
		return err
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()

	rawdb.WriteBlock(bc.db, genesis)

	bc.genesisBlock = genesis
	bc.insert(bc.genesisBlock)

	return nil
}

// SetHead rewinds the local chain to a new head. In the case of headers, everything
// above the new head will be deleted and the new one set.
func (bc *BlockChain) SetHead(head uint64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for block := bc.CurrentBlock(); block.Number().Uint64() >= head; {
		rawdb.DeleteBlockHash(bc.db, block.Number().Uint64())
		rawdb.DeleteBlock(bc.db, block.Hash(), block.Number().Uint64())
		bc.currentBlock.Store(bc.GetBlock(block.ParentHash(), block.Number().Uint64()-1))
	}
	// If either blocks reached nil, reset to the genesis state
	if currentBlock := bc.CurrentBlock(); currentBlock == nil {
		bc.currentBlock.Store(bc.genesisBlock)
	}

	rawdb.WriteHeadBlockHash(bc.db, bc.CurrentBlock().Hash())
	return nil
}
