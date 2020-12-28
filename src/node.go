package kagi

const (
	NodeSize   int16 = 8192
	IntSize    int8  = 4
	FlagSize   int8  = 1
	HeaderSize       = 8170 // 8192 - (2 * 1) - (4 * 5)
)

type Header struct {
	degree     int32
	rootOffset uint32
}

type Node struct {
	// On disk data
	// Flags
	isRoot    bool
	isDeleted bool

	// Counts
	numChildren uint32

	// Offsets
	offset           uint32
	parentOffset     uint32
	leftChildOffset  uint32
	rightChildOffset uint32

	// key
	keySize uint32
	key     string

	// value
	// represented as Leaf in memory
	leaf *Leaf
}

type Leaf struct {
	// On disk value
	value     string
	valueSize uint32

	// space left in node after value
	freeSpace uint32
}

func NewLeaf(v string, keySize int) *Leaf {
	l := &Leaf{}

	l.value = v
	l.valueSize = len(v)
	l.freeSpace = NodeSize - HeaderSize - keySize - l.valueSize

	return l
}

func NodeFromBytes(b []byte) *Node {
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
	node.numChildren = IntFromBytes(b[offset:IntSize])
	offset += IntSize

	// key
	node.keySize = IntFromBytes(b[offset:IntSize])
	offset += IntSize
	node.key = string(b[offset:node.keySize])
	offset += node.keySize

	// offsets
	node.offset = IntFromBytes(b[offset:IntSize])
	offset += IntSize
	node.parentOffset = IntFromBytes(b[offset:IntSize])
	offset += IntSize
	node.leftChildOffset = IntFromBytes(b[offset:IntSize])
	offset += IntSize
	node.rightChildOffset = IntFromBytes(b[offset:IntSize])
	offset += IntSize

	// adding children offsets
	node.leaf = LeafFromBytes(b[offset:], offset)

	return Node
}

func LeafFromBytes(b []byte, nonLeafOffset int) *Leaf {
	offset := 0
	leaf := &Leaf{}

	leaf.valueSize = IntFromBytes(b[offset:IntSize])
	offset += IntSize
	leaf.value = string(b[offset:leaf.valueSize])
	leaf.freeSpace = NodeSize - nonLeafOffset - leaf.valueSize

	return leaf
}

func (n *Node) toBytes() []byte {
	b := make([]byte, NodeSize)

	// flags
	b = append(b, n.isRoot)
	b = append(b, n.isDeleted)

	// count
	b = append(b, BytesFromInt(n.numChildren))

	// key
	b = append(b, l.key)
	b = append(b, BytesFromInt(l.keySize))

	// offsets
	b = append(b, BytesFromInt(n.offset))
	b = append(b, BytesFromInt(n.parentOffset))
	b = append(b, BytesFromInt(n.leftChildOffset))
	b = append(b, BytesFromInt(n.rightChildOffset))

	// leaf
	if checkHasLeaf(n) {
		b = append(b, l.toBytes())
	}
}

func (l *Leaf) toBytes() []byte {
	b := make([]byte, LeafSize)

	b = append(b, BytesFromInt(l.valueSize))
	b = append(b, l.value)

	return b
}

func checkHasLeaf(n *Node) bool {
	return n.numChildren == 0
}
