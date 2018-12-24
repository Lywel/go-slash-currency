package currency

import (
	"bitbucket.org/ventureslash/go-ibft"
	"errors"
	"os"
	"strconv"
	"time"
)

var (
	errFailedToContactCA = errors.New("Failed to get a starting point from the CA")
)

const (
	blockchainSyncDelay = 10 * time.Second
)

func (c *Currency) mine() {
	if c.isProposer() {
		c.submitBlock()
	}
}

func (c *Currency) setTimer() {
	c.mineTimer = time.AfterFunc(blockInterval, c.mine)
}

func (c *Currency) handleTimeout() {
	c.logger.Warning("Block timeout, next proposer")
	c.currentSigner++
	c.blockTimeout = time.AfterFunc(blockTimeoutTime, c.handleTimeout)
	if c.isProposer() {
		c.mine()
	}
}

func (c *Currency) isProposer() bool {
	i, _ := c.valSet.GetByAddress(c.backend.Address())
	return c.currentSigner%uint64(c.valSet.Size()) == uint64(i)
}

func (c *Currency) getStartingBlockNumber() (uint64, error) {
	i, _ := strconv.ParseInt(os.Getenv("CA_START"), 10, 64)
	return uint64(i), nil
}

func (c *Currency) waitForCAAuthorization() error {
	// Contact CA to get the starting block
	startingBlock, err := c.getStartingBlockNumber()
	if err != nil {
		return errFailedToContactCA
	}
	c.logger.Infof("Got Authorisation to start at block %d", startingBlock)

	// Wait for this currencyBlock.Number >= startingBlock while syncing th bc
	for currentBlock := c.blockchain.CurrentBlock().Number().Uint64(); currentBlock < startingBlock; {
		c.logger.Infof("Waiting for starting block to be created (%d/%d)", currentBlock, startingBlock)
		// Wait a little bit
		time.Sleep(blockchainSyncDelay)
		// Get the last blocks
		c.syncBlockchain()
		currentBlock = c.blockchain.CurrentBlock().Number().Uint64()
	}

	return nil
}

func (c *Currency) isAuthorizedValidator(addr ibft.Address) bool {
	return true
	// TODO: http.Get("<ca url>/peer-authorization/" + addr) -> { addr: addr, startingBlock: 53 }
	// return currentBlock < res.startingBlock
}
