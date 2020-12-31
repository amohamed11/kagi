package kagi

import (
	"encoding/binary"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	TRUE               = 1
	FALSE              = 0
	KEY_NOT_FOUND      = Error("key not found in database.")
	KEY_ALREADY_EXISTS = Error("key already exists in database.")
	ERROR_WRITING_NODE = Error("error writing given node to database file")
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Uint32FromBytes(b []byte) uint32 {
	newInt := binary.LittleEndian.Uint32(b)
	return newInt
}

func BytesFromUint32(i uint32) []byte {
	b := make([]byte, Int32Size)
	binary.LittleEndian.PutUint32(b, i)
	return b
}

func Uint16FromBytes(b []byte) uint16 {
	newInt := binary.LittleEndian.Uint16(b)
	return newInt
}

func BytesFromUint16(i uint16) []byte {
	b := make([]byte, Int16Size)
	binary.LittleEndian.PutUint16(b, i)
	return b
}
