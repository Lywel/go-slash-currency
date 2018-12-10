package types

import (
	"bitbucket.org/ventureslash/go-ibft"
)

const (
	// ReceiptStatusFailed is the status code of a transaction if execution failed.
	ReceiptStatusFailed = uint64(0)
	// ReceiptStatusSuccessful is the status code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

// Receipt represents the results of a transaction
type Receipt struct {
	TxHash ibft.Hash
	Status uint64
}

// Receipts is an array of Receipt
type Receipts []*Receipt

// NewReceipt creates a transaction receipt
func NewReceipt(txHash ibft.Hash, status uint64) *Receipt {
	return &Receipt{
		TxHash: txHash,
		Status: status,
	}
}
