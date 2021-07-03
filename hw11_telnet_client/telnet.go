package main

import (
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClientImpl struct {
	address      string
	timeout      time.Duration
	inputStream  io.ReadCloser
	outputStream io.Writer
	connection   net.Conn
}

func (t *telnetClientImpl) Close() error {
	return t.connection.Close()
}

func (t *telnetClientImpl) Send() error {
	_, err := io.Copy(t.connection, t.inputStream)
	return err
}

func (t *telnetClientImpl) Receive() error {
	_, err := io.Copy(t.outputStream, t.connection)
	return err
}

func (t *telnetClientImpl) Connect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	t.connection = conn
	return nil
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClientImpl{
		address:      address,
		timeout:      timeout,
		inputStream:  in,
		outputStream: out,
	}
}
