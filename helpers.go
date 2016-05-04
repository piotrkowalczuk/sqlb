package sqlb

type counter struct {
	index int64
}

func (c *counter) get() int64 {
	c.inc()
	return c.index
}

func (c *counter) inc() {
	c.index++
}

func (c *counter) reset() {
	c.index = 0
}
