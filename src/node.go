package kagi

const (
	NodeSize   int32 = 4096 // max size of a node
	IntSize    int32 = 4    // size of uint32 used for offsets in node
	FlagSize   int32 = 2    // size of uint16 used for flags in nodes
	HeaderSize int32 = 4068 // 4096 - (2 * 2) - (4 * 6)
)

type Node struct {
	// On disk data
	// Flags
	isRoot    uint16
	isDeleted uint16

	// Counts
	numChildren uint32

	// Offsets
	offset           uint32
	parentOffset     uint32
	leftChildOffset  uint32
	rightChildOffset uint32

	// key
	keySize int32
	key     string

	// value
	// represented as Leaf in memory
	leaf *Leaf
}

type Leaf struct {
	// On disk value
	value     string
	valueSize int32

	// space left in node after value
	freeSpace int32
}

func NewLeaf(v string, keySize int32) *Leaf {
	l := &Leaf{}

	l.value = v
	l.valueSize = int32(len(v))
	l.freeSpace = NodeSize - HeaderSize - keySize - l.valueSize

	return l
}

func NodeFromBytes(b []byte) *Node {
	offset := int32(0)
	node := &Node{}

	// flags
	node.isRoot = Uint16FromBytes(b[offset : offset+FlagSize])
	offset += FlagSize
	node.isDeleted = Uint16FromBytes(b[offset : offset+FlagSize])
	offset += FlagSize

	// count
	node.numChildren = Uint32FromBytes(b[offset : offset+IntSize])
	offset += IntSize

	// key
	node.keySize = int32(Uint32FromBytes(b[offset : offset+IntSize]))
	offset += IntSize
	node.key = string(b[offset : offset+node.keySize])
	offset += node.keySize

	// offsets
	node.offset = Uint32FromBytes(b[offset : offset+IntSize])
	offset += IntSize
	node.parentOffset = Uint32FromBytes(b[offset : offset+IntSize])
	offset += IntSize
	node.leftChildOffset = Uint32FromBytes(b[offset : offset+IntSize])
	offset += IntSize
	node.rightChildOffset = Uint32FromBytes(b[offset : offset+IntSize])
	offset += IntSize

	// adding children offsets
	node.leaf = LeafFromBytes(b[offset:], offset)

	return node
}

func LeafFromBytes(b []byte, nonLeafOffset int32) *Leaf {
	offset := int32(0)
	leaf := &Leaf{}

	leaf.valueSize = int32(Uint32FromBytes(b[offset : offset+IntSize]))
	offset += IntSize
	leaf.value = string(b[offset : offset+leaf.valueSize])
	leaf.freeSpace = NodeSize - nonLeafOffset - int32(leaf.valueSize)

	return leaf
}

func (n *Node) toBytes() []byte {
	b := make([]byte, NodeSize)
	offset := int32(0)

	// flags
	copy(b[offset:], BytesFromUint16(n.isRoot))
	offset += FlagSize
	copy(b[offset:], BytesFromUint16(n.isDeleted))
	offset += FlagSize

	// count
	copy(b[offset:], BytesFromUint32(n.numChildren))
	offset += IntSize

	// key
	copy(b[offset:], BytesFromUint32(uint32(n.keySize)))
	offset += IntSize
	copy(b[offset:], n.key)
	offset += n.keySize

	// offsets
	copy(b[offset:], BytesFromUint32(n.offset))
	offset += IntSize
	copy(b[offset:], BytesFromUint32(n.parentOffset))
	offset += IntSize
	copy(b[offset:], BytesFromUint32(n.leftChildOffset))
	offset += IntSize
	copy(b[offset:], BytesFromUint32(n.rightChildOffset))
	offset += IntSize

	// leaf
	if checkHasLeaf(n) {
		copy(b[offset:], n.leaf.toBytes(offset))
	}

	return b
}

func (l *Leaf) toBytes(headerOffset int32) []byte {
	size := NodeSize - headerOffset
	b := make([]byte, size)
	offset := int32(0)

	copy(b[offset:], BytesFromUint32(uint32(l.valueSize)))
	offset += IntSize
	copy(b[offset:], l.value)

	return b
}

func checkHasLeaf(n *Node) bool {
	return n.numChildren == 0
}
