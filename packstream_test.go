package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPackStream_SendNull(t *testing.T) {
	var buffer bytes.Buffer
	ps := NewPackStream(&buffer, &buffer)
	_ = ps.SendNull()
	actual, _ := ps.Receive()

	assert.Nil(t, actual)
}

func TestPackStream_SendBoolean(t *testing.T) {
	for _, expected := range []bool{true, false} {
		var buffer bytes.Buffer
		ps := NewPackStream(&buffer, &buffer)
		_ = ps.SendBoolean(expected)
		actual, _ := ps.Receive()

		assert.Equal(t, expected, actual)
	}
}

func TestPackStream_SendInt(t *testing.T) {
	for _, expected := range []int64{
		-9_223_372_036_854_775_808, -2_147_483_649, // INT_64
		-2_147_483_648, -32_769, // INT_32
		-32_768, -129, // INT_16
		-128, -17, // INT_8
		-16, 127, // TINY_INT
		128, 32_767, // INT_16
		32_768, 2_147_483_647, // INT_32
		2_147_483_648, 9_223_372_036_854_775_807, // INT_64
	} {
		var buffer bytes.Buffer
		ps := NewPackStream(&buffer, &buffer)
		_ = ps.SendInt(expected)
		actual, _ := ps.Receive()

		assert.Equal(t, expected, actual)
	}
}

func TestPackStream_SendString(t *testing.T) {
	for _, expected := range []string{
		"",                           // EMPTY
		"A",                          // STRING_16
		"one",                        // STRING_16
		string(make([]byte, 16)),     // STRING_255
		string(make([]byte, 256)),    // STRING_65_535
		string(make([]byte, 65_536)), // STRING_2_147_483_647
	} {
		var buffer bytes.Buffer
		ps := NewPackStream(&buffer, &buffer)
		_ = ps.SendString(expected)
		actual, _ := ps.Receive()

		assert.Equal(t, expected, actual)
	}
}

func TestPackStream_SendDictionary(t *testing.T) {
	for _, expected := range []map[string]interface{}{
		{},                  // A0
		{"one": "eins"},     // A1
		generateMap(16),     // D8
		generateMap(256),    // D9
		generateMap(65_536), // DA
	} {
		var buffer bytes.Buffer
		ps := NewPackStream(&buffer, &buffer)
		_ = ps.SendDictionary(expected)
		actual, _ := ps.Receive()

		assert.Equal(t, expected, actual)
	}
}

func generateMap(size int) map[string]interface{} {
	m := map[string]interface{}{}
	for i := 0; i < size; i++ {
		m[strconv.Itoa(i)] = "value"
	}
	return m
}
