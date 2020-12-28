package kagi

const (
	NodeSize   int32 = 4096
	IntSize    int32 = 4
	FlagSize   int32 = 2
	HeaderSize int32 = 4074 // 4096 - (2 * 1) - (4 * 5)
)

type Header struct {
	degree     int32
	rootOffset uint32
}

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

	// flags
	b = append(b, BytesFromUint16(n.isRoot)...)
	b = append(b, BytesFromUint16(n.isDeleted)...)

	// count
	b = append(b, BytesFromUint32(n.numChildren)...)

	// key
	b = append(b, n.key...)
	b = append(b, BytesFromUint32(uint32(n.keySize))...)

	// offsets
	b = append(b, BytesFromUint32(n.offset)...)
	b = append(b, BytesFromUint32(n.parentOffset)...)
	b = append(b, BytesFromUint32(n.leftChildOffset)...)
	b = append(b, BytesFromUint32(n.rightChildOffset)...)

	// leaf
	if checkHasLeaf(n) {
		b = append(b, n.leaf.toBytes(n.keySize)...)
	}

	return b
}

func (l *Leaf) toBytes(keySize int32) []byte {
	size := NodeSize - int32(keySize)
	b := make([]byte, size)

	b = append(b, BytesFromUint32(uint32(l.valueSize))...)
	b = append(b, l.value...)

	return b
}

func checkHasLeaf(n *Node) bool {
	return n.numChildren == 0
}
