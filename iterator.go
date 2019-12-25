package godb

// Iterator
type Iterator struct {
	current int
	count   int
}

// New Iterator
func NewIterator(len int) *Iterator {
	return &Iterator{
		current: -1,
		count:   len,
	}
}

// Iterator next
func (c *Iterator) Next() bool {
	if c.current >= c.count-1 {
		c.Reset()
		return false
	}
	c.current++
	return true
}

// Get cursor
func (c *Iterator) Cursor() int {
	return c.current
}

// Reset cursor
func (c *Iterator) Reset() *Iterator {
	c.current = -1
	return c
}

// Get count
func (c *Iterator) Count() int {
	return c.count
}

// Set count
func (c *Iterator) SetCount(count int) *Iterator {
	c.count = count
	return c
}



