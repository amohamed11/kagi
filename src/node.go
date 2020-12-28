package kagi

import "encoding/binary"

const (
	NodeSize int16 = 8192
	IntSize  int8  = 4
	FlagSize int8  = 1
	LeafSize int8  = 8174 // 8192 - 1 - 1 - (4*4)
)

type Header struct {
	degree     int32
	rootOffset uint32
}

type Node struct {
	// On Disk
	// Flags
	isRoot    bool
	isDeleted bool

	// Counts
	numChildren uint32

	// Offsets
	parentOffset     uint32
	leftChildOffset  uint32
	rightChildOffset uint32

	// In Memory
	leaf *Leaf
}

type Leaf struct {
	// On Disk
	key       string
	keySize   uint32
	value     string
	valueSize uint32

	// In Memory
	freeSpace uint32
}

func NodefromBytes(b []byte) *Node {
	offset := 0
	node := &Node{}

	// flags
	node.isLeaf = b[offset:FlagSize]
	offset += FlagSize
	node.isRoot = b[offset:FlagSize]
	offset += FlagSize
	node.isDeleted = b[offset:FlagSize]
	offset += FlagSize

	// count
	node.numChildren = intFromBytes(b[offset:IntSize])
	offset += IntSize

	// offsets
	node.parentOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize
	node.leftChildOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize
	node.rightChildOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize

	// adding children offsets
	node.leaf = LeafFromBytes(b[offset:BlockSize])
	return Node
}

func LeafFromBytes(b []byte) *Leaf {
	offset := 0
	leaf := &Leaf{}

	// sizes
	leaf.keySize = intFromBytes(b[offset:IntSize])
	offset += IntSize
	leaf.valueSize = intFromBytes(b[offset:IntSize])
	offset += IntSize
	leaf.freeSpace = (leaf.keySize + leaf.valueSize + (IntSize * 2)) - BlockSize

	// key-value pair
	leaf.key = string(b[offset:leaf.keySize])
	offset += leaf.keySize
	leaf.value = string(b[offset:leaf.valueSize])

	return leaf
}

func BytesFromNode(n *Node) []byte {
	b := make([]byte, BlockSize)
	offset = 0

	// flags
	b = append(b, n.isRoot)
	b = append(b, n.isDeleted)

	// count
	b = append(b, bytesFromInt(n.numChildren))

	// offsets
	b = append(b, bytesFromInt(n.parentOffset))
	b = append(b, bytesFromInt(n.leftChildOffset))
	b = append(b, bytesFromInt(n.rightChildOffset))

	// leaf
	if checkIsLeaf(n) {
		b = append(b, BytesFromLeaf(n.leaf))
	}
}

func BytesFromLeaf(l *Leaf) []byte {
	b := make([]byte, LeafSize)
	offset = 0

	// sizes
	b = append(b, bytesFromInt(l.keySize))
	b = append(b, bytesFromInt(l.valueSize))

	// key-value pair
	b = append(b, l.key)
	b = append(b, l.value)

	return b
}

func checkIsLeaf(n *Node) bool {
	return n.numChildren == 0
}

func intFromBytes(b []byte) uint32 {
	newInt := binary.LittleEndian.Uint32(b)
	return newInt
}

func bytesFromInt(i uint32) []byte {
	b := make([]byte, IntSize)
	binary.LittleEndian.PutUint32(b[0:], i)
	return b
}
