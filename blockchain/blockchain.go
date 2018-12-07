package blockchain

import (
	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/rawdb"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

// BlockChain is the structure managing and storing blocks
type BlockChain struct {
	db           *leveldb.DB
	genesisBlock *types.Block
	currentBlock atomic.Value
	mu           sync.RWMutex // global mutex for locking chain operations
	chainmu      sync.RWMutex // blockchain insertion lock
}

// New resturns a new instance of Blockchain
func New(file string) (*BlockChain, error) {
	db, err := rawdb.InitDB(file)
	if err != nil {
		return nil, err
	}

	bc := &BlockChain{
		db: db,
	}

	bc.genesisBlock = bc.readOrCreateGenesisBlock()

	return bc, nil
}

func (bc *BlockChain) readOrCreateGenesisBlock() *types.Block {
	genesis := bc.GetBlockByNumber(0)
	if genesis == nil {
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

	return genesis
}

// GetBlockByHash retrieves a block from the database by hash
func (bc *BlockChain) GetBlockByHash(hash ibft.Hash) *types.Block {
	number := rawdb.ReadBlockNumber(bc.db, hash)
	if number == nil {
		return nil
	}
	return bc.GetBlock(hash, *number)
}

// GetBlockByNumber retrieves a block from the database by number
func (bc *BlockChain) GetBlockByNumber(number uint64) *types.Block {
	hash := rawdb.ReadBlockHash(bc.db, number)
	if hash == (ibft.Hash{}) {
		return nil
	}
	return bc.GetBlock(hash, number)
}

// GetBlock retrieves a block from the database by hash and number
func (bc *BlockChain) GetBlock(hash ibft.Hash, number uint64) *types.Block {
	block := rawdb.ReadBlock(bc.db, hash, number)
	if block == nil {
		return nil
	}
	return block
}

// WriteBlock writes the block to the database
func (bc *BlockChain) WriteBlock(block *types.Block, receipts []*types.Receipt) error {
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
