package types

const (
	// ReceiptStatusFailed is the status code of a transaction if execution failed.
	ReceiptStatusFailed = uint64(0)
	// ReceiptStatusSuccessful is the status code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

// Receipt represents the results of a transaction
type Receipt struct {
	TxHash []byte
	Status uint64
}

// NewReceipt creates a transaction receipt

func NewReceipt(txHash []byte, status uint64) *Receipt {
	return &Receipt{
		TxHash: txHash,
		Status: status,
	}
}
