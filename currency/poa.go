package currency

var currentSigner = uint64(0)

func (c *Currency) mine() {
	c.submitBlock()
}

func (c *Currency) isProposer() bool {
	i, _ := c.valSet.GetByAddress(c.backend.Address())
	return currentSigner%uint64(c.valSet.Size()) == uint64(i)
}
