package statsd

import (
	"fmt"
	"io"
	"time"
)

type Client struct {
	prefix string
	writer io.Writer
}

func NewClient(prefix string, w io.Writer) (c *Client) {

	if prefix != "" {
		prefix += "."
	}

	c = &Client{
		prefix: prefix,
		writer: w,
	}

	return c
}

// Increment a statsd counter
func (c *Client) Increment(name string, count int64) (err error) {

	err = c.send(name, "%d|c", count)
	if err != nil {
		return err
	}

	return nil
}

// Decrement a statsd counter
func (c *Client) Decrement(name string, count int64) (err error) {

	err = c.Increment(name, -count)
	if err != nil {
		return err
	}

	return nil
}

// Time send a statsd timing
func (c *Client) Time(name string, duration time.Duration) (err error) {

	return c.send(name, "%d|ms", millisecond(duration))
	if err != nil {
		return err
	}

	return nil
}

// Gauge send a tatsd gauge value
func (c *Client) Gauge(name string, value int64) (err error) {

	err = c.send(name, "%d|g", value)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) send(stat string, format string, args ...interface{}) (err error) {

	format = fmt.Sprintf("%s%s:%s\n", c.prefix, stat, format)
	_, err = fmt.Fprintf(c.writer, format, args...)

	return err
}

func millisecond(d time.Duration) int64 {

	return int64(d.Seconds() * 1000)
}
