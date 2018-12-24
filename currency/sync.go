package currency

import (
	"io/ioutil"
	"net/http"

	"bitbucket.org/ventureslash/go-slash-currency/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func (c *Currency) syncBlockchain() {
	for _, remote := range c.remotes {
		c.logger.Info("Syncing blockchain from: ", remote)
		// TODO: only request the missing blocks
		resp, err := http.Get("http://" + remote + "/state")
		if err != nil {
			c.logger.Warningf("failed to get state from %s: %v", remote, err)
			continue
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.logger.Warningf("failed to read state from %s: %v", remote, err)
			continue
		}
		state := &struct {
			Blockchain   []*types.Block
			Transactions []*types.Transaction
		}{}

		err = rlp.DecodeBytes(body, state)
		if err != nil {
			c.logger.Warningf("failed to decode state from %s: %v", remote, err)
			continue
		}
		err = c.blockchain.InsertChain(state.Blockchain[c.blockchain.CurrentBlock().Number().Uint64()+1:])
		if err != nil {
			c.logger.Warningf("failed to insert blockchain from %s: %v", remote, err)
			continue
		}

		// No error triggered a continue. The blockchain is synchronized
		return
	}
}
