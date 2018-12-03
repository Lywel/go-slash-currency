package currency

import (
	"bitbucket.org/ventureslash/go-slash-currency/types"
)

func (c *Currency) handleState(blockchain []*types.Block, transactions types.Transactions) {
	c.logger.Info("Blockchain and Txs fetched:")
	c.logger.Info("blockchain: ", blockchain)
	c.logger.Info("transactions: ", transactions)
	c.blockchain = blockchain
	c.transactions = types.TxDifference(c.transactions, transactions)
	c.transactions = append(transactions, c.transactions...)
}
