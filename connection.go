package main

import (
	"fmt"
	"net"
)

type Connection interface {
	Connect() (net.Conn, error)
}

type connection struct {
	address string
}

func NewConnection(address string) Connection {
	return &connection{address}
}

func (c *connection) Connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, err
	}
	// Handshake: Bolt identification and version request
	handshake := []byte{0x60, 0x60, 0xB0, 0x17, // identification
		0x00, 0x03, 0x03, 0x04, //
		0x00, 0x00, 0x01, 0x04, //
		0x00, 0x00, 0x00, 0x04, //
		0x00, 0x00, 0x00, 0x03, //
	}
	if _, err := conn.Write(handshake); err != nil {
		return nil, err
	}
	// Receive the version response
	response := make([]byte, 4)
	if _, err := conn.Read(response); err != nil {
		return nil, err
	}
	fmt.Printf("server version response [%d.%d]\n", response[3], response[2])

	return conn, nil
}
