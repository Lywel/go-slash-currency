package currency

import "time"

var currentSigner = uint64(0)

func (c *Currency) mine() {
	c.submitBlock()
}

func (c *Currency) setTimer() {
	c.mineTimer = time.AfterFunc(blockInterval, c.mine)
}

func (c *Currency) handleTimeout() {
	c.logger.Warning("Block timeout, next proposer")
	currentSigner++
	c.blockTimeout = time.AfterFunc(blockTimeoutTime, c.handleTimeout)
	if c.isProposer() {
		c.mine()
	}
}

func (c *Currency) isProposer() bool {
	i, _ := c.valSet.GetByAddress(c.backend.Address())
	return currentSigner%uint64(c.valSet.Size()) == uint64(i)
}
