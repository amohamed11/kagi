package kagi

func NodeFromBytes(b []byte) *Node {
	offset := int32(0)
	node := &Node{}

	// flags
	node.isRoot = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size
	node.isDeleted = Uint16FromBytes(b[offset : offset+Int16Size])
	offset += Int16Size

	// count
	node.numKeys = Uint16FromBytes(b[offset : offset+Int32Size])
	offset += Int16Size
	node.numChildren = Uint16FromBytes(b[offset : offset+Int32Size])
	offset += Int16Size
	node.numLeaves = Uint16FromBytes(b[offset : offset+Int32Size])
	offset += Int16Size

	// offsets
	node.offset = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size
	node.parentOffset = Uint32FromBytes(b[offset : offset+Int32Size])
	offset += Int32Size

	// children offsets
	for i := 0; uint16(i) < node.numChildren; i++ {
		node.childOffsets[i] = Uint32FromBytes(b[offset : offset+Int32Size])
		offset += Int32Size
	}

	// keys
	if !node.checkHasLeaf() {
		node.keys = make([]*Data, node.numKeys)
		for i := 0; uint16(i) < node.numKeys; i++ {
			node.keys[i].size = int32(Uint32FromBytes(b[offset : offset+Int32Size]))
			offset += Int32Size
			node.keys[i].data = b[offset : offset+node.keys[i].size]
			offset += node.keys[i].size
		}
	} else {
		node.leaves = make([]*Leaf, node.numLeaves)
		leafOffset := int32(0)
		for i := 0; uint16(i) < node.numLeaves; i++ {
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
	leaf.key.size = int32(Uint32FromBytes(b[offset : offset+Int32Size]))
	offset += Int32Size
	leaf.key.data = b[offset : offset+leaf.key.size]
	offset += leaf.key.size

	// value
	leaf.value = &Data{}
	leaf.value.size = int32(Uint32FromBytes(b[offset : offset+Int32Size]))
	offset += Int32Size
	leaf.value.data = b[offset : offset+leaf.value.size]
	offset += leaf.value.size

	return leaf, offset
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
	copy(b[offset:], BytesFromUint16(n.numKeys))
	offset += Int16Size
	copy(b[offset:], BytesFromUint16(n.numChildren))
	offset += Int16Size
	copy(b[offset:], BytesFromUint16(n.numLeaves))
	offset += Int16Size

	// offsets
	copy(b[offset:], BytesFromUint32(n.offset))
	offset += Int32Size
	copy(b[offset:], BytesFromUint32(n.parentOffset))
	offset += Int32Size

	// children offsets
	for i := 0; i < int(Order); i++ {
		if uint16(i) < n.numChildren {
			copy(b[offset:], BytesFromUint32(n.childOffsets[i]))
		} else {
			copy(b[offset:], BytesFromUint32(uint32(0)))
		}
		offset += Int32Size
	}

	if !n.checkHasLeaf() {
		// keys
		for i := 0; uint16(i) < n.numKeys; i++ {
			copy(b[offset:], BytesFromUint32(uint32(n.keys[i].size)))
			offset += Int32Size
			copy(b[offset:], n.keys[i].data)
			offset += n.keys[i].size
		}
	} else {
		// leaves
		for i := 0; uint16(i) < n.numLeaves; i++ {
			copy(b[offset:], n.leaves[i].toBytes(offset))
		}
	}

	return b
}

func (l *Leaf) toBytes(headerOffset int32) []byte {
	size := BlockSize - headerOffset
	b := make([]byte, size)
	offset := int32(0)

	// key
	copy(b[offset:], BytesFromUint32(uint32(l.key.size)))
	offset += Int32Size
	copy(b[offset:], l.key.data)
	offset += l.key.size

	// value
	copy(b[offset:], BytesFromUint32(uint32(l.value.size)))
	offset += Int32Size
	copy(b[offset:], l.value.data)

	return b
}
