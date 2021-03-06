package bdm

import (
	"io"

	"github.com/tarm/serial"
)

// Client is RS-232C client.
type Client struct {
	serial io.ReadWriteCloser
}

var fixedHeader = []byte{
	0xA6, // Header
	0x01, // monitor ID
	0x00, // Category
	0x00, // Page
	0x00, // Function
}

const control byte = 0x01

// New open a serial port and returns a new client.
func New(port string, baud int) (*Client, error) {
	conf := &serial.Config{
		Name: port,
		Baud: baud,
	}
	s, err := serial.OpenPort(conf)
	if err != nil {
		return nil, err
	}

	c := &Client{
		serial: s,
	}
	return c, nil
}

// Send writes data to the display, and reads result from the display.
func (c *Client) Send(data []byte) (Result, error) {
	_, err := c.write(data)
	if err != nil {
		return nil, err
	}

	return c.read()
}

// Close closes a serial port.
func (c *Client) Close() error {
	return c.serial.Close()
}

func (c *Client) write(data []byte) (int, error) {
	msg := c.build(data)
	return c.serial.Write(msg)
}

// [Header] [MonitorID] [Category] [Page] [Length] [Control] [Command] [Data 0] ... [Data N] [Checksum]
// Or
// [Header] [MonitorID] [Category] [Page] [Length] [Control] [Data 0] [Status] [Checksum]
func (c *Client) read() (Result, error) {
	header := make([]byte, 6)
	_, err := io.ReadFull(c.serial, header)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, header[4]-1)
	_, err = io.ReadFull(c.serial, buf)
	if err != nil {
		return nil, err
	}

	res := Result(append(header, buf...))

	return res, res.CheckChecksum()
}

// [Header] [MonitorID] [Category] [Page] [Function] [Length] [Control] [Data 0] ... [Data N] [Checksum]
func (c *Client) build(data []byte) []byte {
	res := make([]byte, 0, len(data)+8)
	res = append(res, fixedHeader...)
	length := len(data) + 2
	res = append(res, byte(length))
	res = append(res, control)
	res = append(res, data...)

	checksum := checksum(res)
	res = append(res, checksum)

	return res
}

// b[0] xor b[1] xor ... b[n]
func checksum(b []byte) byte {
	res := byte(0)
	for _, v := range b {
		res ^= v
	}
	return res
}
