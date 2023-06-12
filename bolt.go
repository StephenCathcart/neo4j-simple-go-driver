package main

import (
	"errors"
	"net"
)

type Auth struct {
	scheme      string
	principal   string
	credentials string
}

type Bolt interface {
	Hello(auth *Auth) error
	Run() error
	Pull() error
	Query(cypher string) error
	Close() error
}

type bolt struct {
	userAgent string
	conn      net.Conn
	ps        PackStream
}

func NewBolt(address string) (Bolt, error) {
	connection := NewConnection(address)
	packStream := NewPackStream(nil, nil)
	conn, err := connection.Connect()
	return &bolt{userAgent: "neo4j-go", conn: conn, ps: packStream}, err
}

func (b *bolt) Hello(auth *Auth) error {
	token := map[string]interface{}{
		"scheme":      auth.scheme,
		"principal":   auth.principal,
		"credentials": auth.credentials,
	}

	// TODO pack(0x01, auth) -> connection.write(packed)

	return b.ps.SendDictionary(token)
}

func (b *bolt) Run() error {
	return errors.New("not implemented")
}

func (b *bolt) Pull() error {
	return errors.New("not implemented")
}

func (b *bolt) Query(_ string) error {
	return errors.New("not implemented")
}

func (b *bolt) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
