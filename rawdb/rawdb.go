package rawdb

import (
	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"log"
)

// InitDB retrieves a database from path
func InitDB(file string) (*leveldb.DB, error) {
	opt := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}
	// Open the db and recover any potential corruptions
	db, err := leveldb.OpenFile(file, opt)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ReadBlockHash retrieves the hash assigned to a block number.
func ReadBlockHash(db *leveldb.DB, number uint64) ibft.Hash {
	data, _ := db.Get(blockHashKey(number), nil)
	if len(data) == 0 {
		return ibft.Hash{}
	}
	return ibft.BytesToHash(data)
}

// WriteBlockHash stores the hash assigned to a block number.
func WriteBlockHash(db *leveldb.DB, hash ibft.Hash, number uint64) {
	if err := db.Put(blockHashKey(number), hash.Bytes(), nil); err != nil {
		log.Println("Failed to store number to hash mapping", "err", err)
	}
}

// DeleteBlockHash removes the number to hash mapping.
func DeleteBlockHash(db *leveldb.DB, number uint64) {
	if err := db.Delete(blockHashKey(number), nil); err != nil {
		log.Println("Failed to delete number to hash mapping", "err", err)
	}
}

// ReadBlockNumber returns the header number assigned to a hash.
func ReadBlockNumber(db *leveldb.DB, hash ibft.Hash) *uint64 {
	data, _ := db.Get(blockNumberKey(hash), nil)
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

// WriteHeadBlockHash stores the head block's hash.
func WriteHeadBlockHash(db *leveldb.DB, hash ibft.Hash) {
	if err := db.Put(headBlockKey, hash.Bytes(), nil); err != nil {
		log.Println("Failed to store last block's hash", "err", err)
	}
}

// ReadHeadBlockHash stores the head block's hash.
func ReadHeadBlockHash(db *leveldb.DB) ibft.Hash {
	data, _ := db.Get(headBlockKey, nil)
	if len(data) == 0 {
		return ibft.Hash{}
	}
	return ibft.BytesToHash(data)
}

// ReadBlockRLP retrieves a full block (headaer and transactions) in RLP encoding.
func ReadBlockRLP(db *leveldb.DB, hash ibft.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(blockKey(number, hash), nil)
	return data
}

// ReadBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
func ReadBlock(db *leveldb.DB, hash ibft.Hash, number uint64) *types.Block {
	data := ReadBlockRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	block := &types.Block{}
	if err := rlp.Decode(bytes.NewReader(data), block); err != nil {
		log.Println("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return block
}

// WriteBlockRLP stores an RLP encoded block into the database.
func WriteBlockRLP(db *leveldb.DB, hash ibft.Hash, number uint64, rlp rlp.RawValue) {
	// Write the hash -> number mapping
	var encoded = encodeBlockNumber(number)
	key := blockNumberKey(hash)
	if err := db.Put(key, encoded, nil); err != nil {
		log.Println("Failed to store hash to number mapping", "err", err)
	}
	if err := db.Put(blockKey(number, hash), rlp, nil); err != nil {
		log.Println("Failed to store block body", "err", err)
	}
}

// WriteBlock serializes a block into the database, header and body separately.
func WriteBlock(db *leveldb.DB, block *types.Block) {
	data, err := rlp.EncodeToBytes(block)
	if err != nil {
		log.Println("Failed to RLP encode body", "err", err)
	}
	WriteBlockRLP(db, block.Hash(), block.Number().Uint64(), data)
}

// DeleteBlock removes all block data associated with a hash.
func DeleteBlock(db *leveldb.DB, hash ibft.Hash, number uint64) {
	DeleteReceipts(db, hash, number)
	if err := db.Delete(blockNumberKey(hash), nil); err != nil {
		log.Println("Failed to delete hash to number mapping", "err", err)
	}
	if err := db.Delete(blockKey(number, hash), nil); err != nil {
		log.Println("Failed to delete block", "err", err)
	}
}

// HasBlock verifies the existence of a block corresponding to the hash.
func HasBlock(db *leveldb.DB, hash ibft.Hash, number uint64) bool {
	if has, err := db.Has(blockKey(number, hash), nil); !has || err != nil {
		return false
	}
	return true
}

// ReadReceipts retrieves all the transaction receipts belonging to a block.
func ReadReceipts(db *leveldb.DB, hash ibft.Hash, number uint64) types.Receipts {
	// Retrieve the flattened receipt slice
	data, _ := db.Get(blockReceiptsKey(number, hash), nil)
	if len(data) == 0 {
		return nil
	}

	receipts := types.Receipts{}
	if err := rlp.DecodeBytes(data, &receipts); err != nil {
		log.Println("Invalid receipt array RLP", "hash", hash, "err", err)
		return nil
	}
	return receipts
}

// WriteReceipts stores all the transaction receipts belonging to a block.
func WriteReceipts(db *leveldb.DB, hash ibft.Hash, number uint64, receipts types.Receipts) {
	bytes, err := rlp.EncodeToBytes(receipts)
	if err != nil {
		log.Println("Failed to encode block receipts", "err", err)
	}
	// Store the flattened receipt slice
	if err := db.Put(blockReceiptsKey(number, hash), bytes, nil); err != nil {
		log.Println("Failed to store block receipts", "err", err)
	}
}

// DeleteReceipts removes all receipt data associated with a block hash.
func DeleteReceipts(db *leveldb.DB, hash ibft.Hash, number uint64) {
	if err := db.Delete(blockReceiptsKey(number, hash), nil); err != nil {
		log.Println("Failed to delete block receipts", "err", err)
	}
}
