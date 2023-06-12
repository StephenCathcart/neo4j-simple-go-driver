package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type PackStream interface {
	Begin() error
	End() error
	SendNull() error
	SendBoolean(value bool) error
	SendInt(value int64) error
	SendString(value string) error
	SendDictionary(values map[string]interface{}) error
	SendValue(value interface{}) error
	Receive() (interface{}, error)
}

type packStream struct {
	writer io.Writer
	reader io.Reader
}

func NewPackStream(writer io.Writer, reader io.Reader) PackStream {
	return &packStream{
		writer: writer,
		reader: reader,
	}
}

func (ps *packStream) Begin() error {
	return errors.New("not implemented")
}

func (ps *packStream) End() error {
	return errors.New("not implemented")
}

func (ps *packStream) SendNull() error {
	_, err := ps.writer.Write([]byte{0xC0})
	return err
}

func (ps *packStream) SendBoolean(b bool) error {
	if b {
		_, err := ps.writer.Write([]byte{0xC3})
		return err
	}
	_, err := ps.writer.Write([]byte{0xC2})
	return err
}

func (ps *packStream) SendInt(i int64) error {
	var err error
	switch {
	case int64(-0x10) <= i && i < int64(0x80):
		_, err = ps.writer.Write([]byte{byte(i)})
	case int64(-0x80) <= i && i < int64(-0x10):
		_, err = ps.writer.Write([]byte{0xC8, byte(i)})
	case int64(-0x8000) <= i && i < int64(0x8000):
		buf := [3]byte{0xC9}
		binary.BigEndian.PutUint16(buf[1:], uint16(i))
		_, err = ps.writer.Write(buf[:])
	case int64(-0x80000000) <= i && i < int64(0x80000000):
		buf := [5]byte{0xCA}
		binary.BigEndian.PutUint32(buf[1:], uint32(i))
		_, err = ps.writer.Write(buf[:])
	default:
		buf := [9]byte{0xCB}
		binary.BigEndian.PutUint64(buf[1:], uint64(i))
		_, err = ps.writer.Write(buf[:])
	}
	return err
}

func (ps *packStream) SendString(value string) error {
	length := len(value)
	var header []byte

	if length <= 15 {
		header = []byte{byte(0x80 | length)}
	} else if length <= 255 {
		header = []byte{0xD0, byte(length)}
	} else if length <= 65_535 {
		header = []byte{0xD1}
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(length))
		header = append(header, buf...)
	} else {
		header = []byte{0xD2}
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(length))
		header = append(header, buf...)
	}
	_, err := ps.writer.Write(append(header, []byte(value)...))
	return err
}

func (ps *packStream) SendDictionary(values map[string]interface{}) error {
	err := ps.sendDictionaryMarker(len(values))
	if err != nil {
		return err
	}

	for key, value := range values {
		err := ps.SendString(key)
		if err != nil {
			return err
		}

		err = ps.SendValue(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *packStream) sendDictionaryMarker(size int) error {
	var marker []byte

	if size <= 15 {
		marker = []byte{byte(0xA0 | size)}
	} else if size <= 255 {
		marker = []byte{0xD8, byte(size)}
	} else if size <= 65_535 {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(size))
		marker = append([]byte{0xD9}, buf...)
	} else {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(size))
		marker = append([]byte{0xDA}, buf...)
	}
	_, err := ps.writer.Write(marker)
	return err
}

func (ps *packStream) SendValue(value interface{}) error {
	switch v := value.(type) {
	case nil:
		return ps.SendNull()
	case bool:
		return ps.SendBoolean(v)
	case int:
		return ps.SendInt(int64(v))
	case int64:
		return ps.SendInt(v)
	case string:
		return ps.SendString(v)
	case map[string]interface{}:
		return ps.SendDictionary(v)
	default:
		return fmt.Errorf("unsupported PackStream value type: %T", value)
	}
}

func (ps *packStream) Receive() (interface{}, error) {
	header := make([]byte, 1)
	_, err := ps.reader.Read(header)
	if err != nil {
		return nil, err
	}

	switch {
	case header[0] == 0xC0: // NULL
		return nil, nil
	case header[0] == 0xC2: // BOOLEAN_FALSE
		return false, nil
	case header[0] == 0xC3: // BOOLEAN_TRUE
		return true, nil
	case header[0]>>4 >= 0xF || header[0]>>4 <= 0x7: // TINY_INT
		return int64(int8(header[0])), nil
	case header[0] == 0xC8: // INT_8
		value, err := ps.readInteger(1)
		return int64(int8(value)), err
	case header[0] == 0xC9: // INT_16
		value, err := ps.readInteger(2)
		return int64(int16(value)), err
	case header[0] == 0xCA: // INT_32
		value, err := ps.readInteger(4)
		return int64(int32(value)), err
	case header[0] == 0xCB: // INT_64
		return ps.readInteger(8)
	case header[0]>>4 == 0x8: // STRING_16
		length := 0x80 ^ header[0]
		return ps.readString(int(length))
	case header[0] == 0xD0: // STRING_255
		length, _ := ps.readInteger(1)
		return ps.readString(int(length))
	case header[0] == 0xD1: // STRING_65_535
		length, _ := ps.readInteger(2)
		return ps.readString(int(length))
	case header[0] == 0xD2: // STRING_2_147_483_647
		length, _ := ps.readInteger(4)
		return ps.readString(int(length))
	case header[0]>>4 == 0xA: // DICTIONARY_16
		return ps.readDictionary(int(0xA0 ^ header[0]))
	case header[0] == 0xD8: // DICTIONARY_255
		size, _ := ps.readInteger(1)
		return ps.readDictionary(int(size))
	case header[0] == 0xD9: // DICTIONARY_65_535
		size, _ := ps.readInteger(2)
		return ps.readDictionary(int(size))
	case header[0] == 0xDA: // DICTIONARY_2_147_483_647
		size, _ := ps.readInteger(4)
		return ps.readDictionary(int(size))
	default:
		return nil, fmt.Errorf("unexpected PackStream type: 0x%02x", header[0])
	}
}

func (ps *packStream) readInteger(size int) (int64, error) {
	buf := make([]byte, size)
	_, err := ps.reader.Read(buf)
	if err != nil {
		return 0, err
	}

	var value int64
	for _, b := range buf {
		value = (value << 8) | int64(b)
	}
	return value, nil
}

func (ps *packStream) readString(length int) (string, error) {
	buf := make([]byte, length)
	_, err := ps.reader.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (ps *packStream) readDictionary(size int) (map[string]interface{}, error) {
	dictionary := make(map[string]interface{})
	for i := 0; i < size; i++ {
		key, err := ps.Receive()
		if err != nil {
			return nil, err
		}

		value, err := ps.Receive()
		if err != nil {
			return nil, err
		}
		dictionary[key.(string)] = value
	}

	return dictionary, nil
}
