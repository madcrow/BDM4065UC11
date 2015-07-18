package bdm

import (
	"fmt"
	"io"

	"github.com/tarm/serial"
)

type Client struct {
	serial *serial.Port
}

var fixedHeader = []byte{
	0xA6, // Header
	0x01, // monitor ID
	0x00, // Category
	0x00, // Page
	0x00, // Function
}

var CheckSumError = fmt.Errorf("Checksum error")

const control byte = 0x01

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

func (c *Client) Send(data []byte) ([]byte, error) {
	_, err := c.write(data)
	if err != nil {
		return nil, err
	}
	resData, err := c.read()
	if err != nil {
		return nil, err
	}

	return resData, nil
}

func (c *Client) Close() error {
	return c.serial.Close()
}

func (c *Client) write(data []byte) (int, error) {
	msg := c.build(data)
	return c.serial.Write(msg)
}

func (c *Client) read() ([]byte, error) {
	header := make([]byte, 5)
	_, err := io.ReadFull(c.serial, header)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, header[4])
	_, err = io.ReadFull(c.serial, buf)
	if err != nil {
		return nil, err
	}

	data := buf[:len(buf)-1]
	checksum := c.checksum(header) ^ c.checksum(data)

	if checksum != buf[len(buf)-1] {
		return nil, CheckSumError
	}

	return append(header, buf...), nil
}

func (c *Client) build(data []byte) []byte {
	res := make([]byte, 0, len(data)+8)
	res = append(res, fixedHeader...)
	length := len(data) + 2
	res = append(res, byte(length))
	res = append(res, control)
	res = append(res, data...)

	checksum := c.checksum(res)
	res = append(res, checksum)

	return res
}

func (_ *Client) checksum(b []byte) byte {
	res := byte(0)
	for _, v := range b {
		res ^= v
	}
	return res
}

var _ io.Closer = &Client{}