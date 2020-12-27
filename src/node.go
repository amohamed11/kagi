package kagi

import "encoding/binary"

const (
	BlockSize int16 = 8192
	IntSize   int8  = 4
	FlagSize  int8  = 1
)

type Header struct {
	degree     int32
	rootOffset uint32
}

type BranchNode struct {
	// Flags
	isLeaf    bool
	isRoot    bool
	isDeleted bool

	// Counts
	numLeafs    uint32
	numChildren uint32

	// Offsets
	parentOffset    uint32
	leftNodeOffset  uint32
	rightNodeOffset uint32
	childOffsets    []uint32
}

type LeafNode struct {
	isLeaf    bool
	currSize  uint32
	key       string
	keySize   uint32
	value     string
	valueSize uint32
}

func BranchNodefromBytes(b []byte) *BranchNode {
	offset := 0
	branchNode := &BranchNode{}

	// flags
	branchNode.isLeaf = b[offset:FlagSize]
	offset += FlagSize
	branchNode.isRoot = b[offset:FlagSize]
	offset += FlagSize
	branchNode.isDeleted = b[offset:FlagSize]
	offset += FlagSize

	// count
	branchNode.numLeafs = intFromBytes(b[offset:IntSize])
	offset += IntSize
	branchNode.numChildren = intFromBytes(b[offset:IntSize])
	offset += IntSize

	// offsets
	branchNode.parentOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize
	branchNode.leftNodeOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize
	branchNode.rightNodeOffset = intFromBytes(b[offset:IntSize])
	offset += IntSize

	// adding children offsets
	branchNode.childOffsets = make([]uint32, branchNode.numChildren)
	for i := 0; i < branchNode.numChildren; i++ {
		branchNode.chilOffsets[i] = intFromBytes(b[offset:IntSize])
		offset += IntSize
	}

	return branchNode
}

func LeafNodefromBytes(b []byte) *LeafNode {
	offset := 0
	leafNode := &LeafNode{}

	// flags
	leafNode.isLeaf = b[offset:FlagSize]
	offset += FlagSize

	// sizes
	leafNode.currSize = intFromBytes(b[offset:IntSize])
	offset += IntSize
	leafNode.keySize = intFromBytes(b[offset:IntSize])
	offset += IntSize
	leafNode.valueSize = intFromBytes(b[offset:IntSize])
	offset += IntSize

	// key-value pair
	leafNode.key = string(b[offset:leafNode.keySize])
	offset += leafNode.keySize
	leafNode.value = string(b[offset:leafNode.valueSize])

	return leafNode
}

func checkIfLeaf(b []byte) bool {
	isLeaf = b[0:FlagSize]
	return isLeaf
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
