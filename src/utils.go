package kagi

import "encoding/binary"

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func IntFromBytes(b []byte) uint32 {
	newInt := binary.LittleEndian.Uint32(b)
	return newInt
}

func BytesFromInt(i uint32) []byte {
	b := make([]byte, IntSize)
	binary.LittleEndian.PutUint32(b[0:], i)
	return b
}
