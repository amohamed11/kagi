package kagi

import (
	"bytes"
	"encoding/binary"
)

func NodeFromBytes(b []byte) *Node {
	offset := int32(0)
	node := &Node{}

	// flags
	node.isRoot = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size
	node.isDeleted = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size

	// count
	node.dbCount = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size
	node.numKeys = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size
	node.numLeaves = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size

	// offsets
	node.offset = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size
	node.parentOffset = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size

	// keys
	if !node.checkHasLeaf() {
		// children offsets
		node.childOffsets = make([]uint32, node.numKeys+1)
		for i := 0; i < int(node.numKeys)+1; i++ {
			node.childOffsets[i] = Uint32FromBytes(b[offset : offset+Int32Size])
			offset += Int32Size
		}

		node.keys = make([][]byte, node.numKeys)
		for i := 0; i < int(node.numKeys); i++ {
			node.keys[i] = bytes.Trim(b[offset:offset+MaxKeySize], "\x00")
			offset += MaxKeySize
		}
	} else {
		node.leaves = make([]*Leaf, node.numLeaves)
		for i := 0; i < int(node.numLeaves); i++ {
			node.leaves[i] = LeafFromBytes(b[offset:])
			offset += MaxKeySize + MaxValueSize
		}
	}

	return node
}

func LeafFromBytes(b []byte) *Leaf {
	offset := int32(0)
	leaf := &Leaf{}

	// key
	leaf.key = bytes.Trim(b[offset:offset+MaxKeySize], "\x00")
	offset += MaxKeySize

	// value
	leaf.value = bytes.Trim(b[offset:offset+MaxValueSize], "\x00")
	offset += MaxValueSize

	return leaf
}

func (n *Node) toBytes() []byte {
	b := make([]byte, BlockSize)
	offset := int32(0)

	// flags
	copy(b[offset:], BytesFromUint16(n.isRoot))
	offset += Int16Size
	copy(b[offset:], BytesFromUint16(n.isDeleted))
	offset += Int16Size

	// count
	copy(b[offset:], BytesFromUint32(n.dbCount))
	offset += Int32Size
	copy(b[offset:], BytesFromUint16(n.numKeys))
	offset += Int16Size
	copy(b[offset:], BytesFromUint16(n.numLeaves))
	offset += Int16Size

	// offsets
	copy(b[offset:], BytesFromUint32(n.offset))
	offset += Int32Size
	copy(b[offset:], BytesFromUint32(n.parentOffset))
	offset += Int32Size

	if !n.checkHasLeaf() {
		// children offsets
		for i := 0; i < int(n.numKeys)+1; i++ {
			copy(b[offset:], BytesFromUint32(n.childOffsets[i]))
			offset += Int32Size
		}

		// keys
		for i := 0; i < int(n.numKeys); i++ {
			copy(b[offset:], n.keys[i])
			offset += MaxKeySize
		}
	} else {
		// leaves
		for i := 0; i < int(n.numLeaves); i++ {
			copy(b[offset:], n.leaves[i].toBytes())
			offset += MaxKeySize + MaxValueSize
		}
	}

	return b
}

func (l *Leaf) toBytes() []byte {
	b := make([]byte, MaxKeySize+MaxValueSize)
	offset := int32(0)

	// key
	copy(b[offset:], l.key)
	offset += MaxKeySize

	// value
	copy(b[offset:], l.value)
	offset += MaxValueSize

	return b
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
