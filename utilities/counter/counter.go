package counter

import "strconv"

type Counter struct {
	i int
}

func (c *Counter) Next() int {
	c.i += 1
	return c.i
}

func (c *Counter) NextString() string {
	num := c.Next()

	return strconv.Itoa(num)
}
