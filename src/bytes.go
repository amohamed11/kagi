package kagi

import (
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

		node.keys = make([]*Data, node.numKeys)
		for i := 0; i < int(node.numKeys); i++ {
			node.keys[i] = &Data{}
			node.keys[i].size = Uint32FromBytes(b[offset : offset+Int32Size])
			offset += Int32Size
			node.keys[i].data = b[offset : offset+int32(node.keys[i].size)]
			offset += int32(node.keys[i].size)
		}
	} else {
		node.leaves = make([]*Leaf, node.numLeaves)
		leafOffset := int32(0)
		for i := 0; i < int(node.numLeaves); i++ {
			node.leaves[i], leafOffset = LeafFromBytes(b[offset+leafOffset:])
		}
	}

	return node
}

func LeafFromBytes(b []byte) (*Leaf, int32) {
	offset := int32(0)
	leaf := &Leaf{}

	// key
	leaf.key = &Data{}
	leaf.key.size = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size
	leaf.key.data = b[offset : offset+int32(leaf.key.size)]
	offset += int32(leaf.key.size)

	// value
	leaf.value = &Data{}
	leaf.value.size = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size
	leaf.value.data = b[offset : offset+int32(leaf.value.size)]
	offset += int32(leaf.value.size)

	return leaf, offset
}

func (n *Node) toBytes() []byte {
	b := make([]byte, 0, BlockSize)
	offset := int32(0)

	// flags
	b = append(b, BytesFromUint16(n.isRoot)...)
	offset += Int16Size
	b = append(b, BytesFromUint16(n.isDeleted)...)
	offset += Int16Size

	// count
	b = append(b, BytesFromUint16(n.numKeys)...)
	offset += Int16Size
	b = append(b, BytesFromUint16(n.numLeaves)...)
	offset += Int16Size

	// offsets
	b = append(b, BytesFromUint32(n.offset)...)
	offset += Int32Size
	b = append(b, BytesFromUint32(n.parentOffset)...)
	offset += Int32Size

	if !n.checkHasLeaf() {
		// children offsets
		for i := 0; i < int(n.numKeys)+1; i++ {
			b = append(b, BytesFromUint32(n.childOffsets[i])...)
			offset += Int32Size
		}

		// keys
		for i := 0; i < int(n.numKeys); i++ {
			b = append(b, BytesFromUint32(n.keys[i].size)...)
			offset += Int32Size
			b = append(b, n.keys[i].data...)
			offset += int32(n.keys[i].size)
		}
	} else {
		// leaves
		for i := 0; i < int(n.numLeaves); i++ {
			leafBytes := n.leaves[i].toBytes()
			b = append(b, leafBytes...)
			offset += int32(len(leafBytes))
		}
	}

	return b[:BlockSize]
}

func (l *Leaf) toBytes() []byte {
	b := make([]byte, 0)
	offset := int32(0)

	// key
	b = append(b, BytesFromUint32(l.key.size)...)
	offset += Int32Size
	b = append(b, l.key.data...)
	offset += int32(l.key.size)

	// value
	b = append(b, BytesFromUint32(l.value.size)...)
	offset += Int32Size
	b = append(b, l.value.data...)
	offset += int32(l.value.size)

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
