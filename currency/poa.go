package currency

import "time"

func (c *Currency) mine() {
	if c.isProposer() {
		c.submitBlock()
	}
	c.mineTimer = time.AfterFunc(blockInterval, c.mine)

}

func (c *Currency) isProposer() bool {
	i, _ := c.valSet.GetByAddress(c.backend.Address())
	lastblock := c.blockchain[len(c.blockchain)-1]
	return lastblock.Number().Uint64()%uint64(c.valSet.Size()) == uint64(i)
}
