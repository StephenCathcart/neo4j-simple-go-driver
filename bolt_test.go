package main

import "testing"

func TestBolt(t *testing.T) {
	t.Skip()
	bolt, _ := NewBolt("127.0.0.1:7687")
	bolt.Hello(&Auth{
		scheme:      "basic",
		principal:   "neo4j",
		credentials: "password",
	})
	bolt.Query("MATCH (n) RETURN n")
	bolt.Close()
}
