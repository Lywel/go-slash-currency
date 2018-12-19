// Package rawdb contains a collection of low level database accessors.
package rawdb

import (
	"bitbucket.org/ventureslash/go-ibft"
	"encoding/binary"
)

// The fields below define the low level database schema prefixing.
var (
	// headBlockKey tracks the latest know full block's hash.
	headBlockKey        = []byte("LastBlock")
	blockPrefix         = []byte("h") // blockPrefix + num (uint64 big endian) + hash -> block
	blockHashSuffix     = []byte("n") // blockPrefix + num (uint64 big endian) + blockHashSuffix -> hash
	blockNumberPrefix   = []byte("H") // blockNumberPrefix + hash -> num (uint64 big endian)
	blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts
	txLookupPrefix      = []byte("l") // txLookupPrefix + hash -> transaction/receipt lookup metadata
)

// TxLookupEntry is a positional metadata to help looking up the data content of
// a transaction or receipt given only its hash.
type TxLookupEntry struct {
	BlockHash  ibft.Hash
	BlockIndex uint64
	Index      uint64
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// BlockHashKey = BlockPrefix + num (uint64 big endian) + BlockHashSuffix
func blockHashKey(number uint64) []byte {
	return append(append(blockPrefix, encodeBlockNumber(number)...), blockHashSuffix...)
}

// headerNumberKey = headerNumberPrefix + hash
func blockNumberKey(hash ibft.Hash) []byte {
	return append(blockNumberPrefix, hash.Bytes()...)
}

// blockKey = blockPrefix + num (uint64 big endian) + hash
func blockKey(number uint64, hash ibft.Hash) []byte {
	return append(append(blockPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// blockReceiptsKey = blockReceiptsPrefix + num (uint64 big endian) + hash
func blockReceiptsKey(number uint64, hash ibft.Hash) []byte {
	return append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// TODO: tx index (advanced)
// txLookupKey = txLookupPrefix + hash
func txLookupKey(hash ibft.Hash) []byte {
	return append(txLookupPrefix, hash.Bytes()...)
}
