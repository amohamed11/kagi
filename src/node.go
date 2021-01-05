package kagi

const (
	Order         int32 = 20   // the upper limit for number of keys that node can hold
	Degree        int32 = 10   // the lower limit for number of keys, (10-20 keys)
	BlockSize     int32 = 4096 // max size of a node
	Int32Size     int32 = 4    // size of uint32 used for offsets in node
	Int16Size     int32 = 2    // size of uint16 used for flags and counts in nodes
	MaxKeySize    int32 = 48
	MaxValueSize  int32 = 144
	NonHeaderSize int32 = 4080 - (Order * 4) // 4096 - flags(2*2) - counts(2*2) - offsets(2*4) - childOffsets(Order*4)
)

type Node struct {
	//----Header----
	// Flags
	isRoot    uint16
	isDeleted uint16

	// Counts
	numKeys   uint16
	numLeaves uint16

	// Offsets
	offset       uint32
	parentOffset uint32
	childOffsets []uint32
	// -------------

	// branching node
	keys [][]byte

	// leaf nodes
	leaves []*Leaf
}

type Leaf struct {
	// data
	key   []byte
	value []byte

	// TODO Calculate in memory
	// freeSpace int32
}

type Data struct {
	data []byte
}

func NewLeaf(k string, v string) *Leaf {
	l := &Leaf{}

	l.key = []byte(k)
	l.value = []byte(v)

	return l
}

// Steps for splitting:
//  1. Splits leaves into 3 parts
//  2. Middle becomes a branching node
//  3. Left & Right become child nodes of the Middle
//  4. Add middle node as child to parent
//  5. If parent is now full, split parent node as well
func (db *DB_CONNECTION) splitNode(fullNode *Node) {
	half := int32(Order / 2)
	leftKey := fullNode.keys[:half]
	middleKey := fullNode.keys[half]
	rightKey := fullNode.keys[half:]

	db.logInfo("splitting branching node")
	db.logInfo("parent node now is key: %s", middleKey)

	// create new node using middle key
	middleBranchNode := &Node{
		isRoot:       fullNode.isRoot,
		numKeys:      1,
		offset:       fullNode.offset,
		parentOffset: fullNode.parentOffset,
	}
	middleBranchNode.keys = make([][]byte, 1)
	middleBranchNode.keys[0] = middleKey

	// offsets for children are at end of the file
	middleBranchNode.childOffsets = make([]uint32, 2)
	middleBranchNode.childOffsets[0] = uint32(BlockSize)*db.count + uint32(Int32Size)
	middleBranchNode.childOffsets[1] = middleBranchNode.childOffsets[0] + uint32(BlockSize)

	// create left & right nodes & link old node's children
	leftChildNode := &Node{
		numKeys:      uint16(half) + 1,
		offset:       middleBranchNode.childOffsets[0],
		parentOffset: middleBranchNode.offset,
		keys:         leftKey,
	}
	leftChildNode.childOffsets = make([]uint32, 0, leftChildNode.numKeys+1)
	leftChildNode.childOffsets = append(leftChildNode.childOffsets, fullNode.childOffsets[:half+1]...)

	rightChildNode := &Node{
		numKeys:      uint16(Order - half),
		offset:       middleBranchNode.childOffsets[1],
		parentOffset: middleBranchNode.offset,
		keys:         rightKey,
	}
	rightChildNode.childOffsets = make([]uint32, 0, rightChildNode.numKeys+1)
	rightChildNode.childOffsets = append(rightChildNode.childOffsets, fullNode.childOffsets[half+1:]...)

	if fullNode.isRoot == TRUE {
		db.writeNodeToFile(middleBranchNode)
	} else {
		// update parent with the middle key
		parent := db.getNodeAt(middleBranchNode.parentOffset)
		parent.addChildNode(db, middleBranchNode)
		db.writeNodeToFile(parent)
	}

	// write newly create nodes
	db.writeNodeToFile(leftChildNode)
	db.writeNodeToFile(rightChildNode)
	db.count += 2
}

func (db *DB_CONNECTION) splitLeaves(fullLeafNode *Node) {
	half := int32(Order / 2)
	leftLeaf := fullLeafNode.leaves[:half]
	middleLeaf := fullLeafNode.leaves[half]
	rightLeaf := fullLeafNode.leaves[half:]

	db.logInfo("splitting leaves")
	db.logInfo("creating new branching node with key: %s", middleLeaf.key)

	// create new node using middle key
	middleBranchNode := &Node{
		isRoot:       fullLeafNode.isRoot,
		numKeys:      1,
		offset:       fullLeafNode.offset,
		parentOffset: fullLeafNode.parentOffset,
	}
	middleBranchNode.keys = make([][]byte, 1)
	middleBranchNode.keys[0] = middleLeaf.key

	// offsets for left & right splits under the middle
	middleBranchNode.childOffsets = make([]uint32, 2)
	middleBranchNode.childOffsets[0] = uint32(BlockSize)*db.count + uint32(Int32Size)
	middleBranchNode.childOffsets[1] = middleBranchNode.childOffsets[0] + uint32(BlockSize)

	// create left & right nodes & populate with split leaves
	leftLeafNode := &Node{
		numLeaves:    uint16(half),
		offset:       middleBranchNode.childOffsets[0],
		parentOffset: middleBranchNode.offset,
		leaves:       leftLeaf,
	}

	rightLeafNode := &Node{
		numLeaves:    uint16(Order - half),
		offset:       middleBranchNode.childOffsets[1],
		parentOffset: middleBranchNode.offset,
		leaves:       rightLeaf,
	}

	// update parent with the middle key
	if fullLeafNode.isRoot == TRUE {
		db.writeNodeToFile(middleBranchNode)
	} else {
		parent := db.getNodeAt(middleBranchNode.parentOffset)
		parent.addChildNode(db, middleBranchNode)
		db.writeNodeToFile(parent)
	}

	// write newly create nodes
	// db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftLeafNode)
	db.writeNodeToFile(rightLeafNode)
	db.count += 1
}

// child node should have a single key
// adds the keys & childoffsets of the node
func (parent *Node) addChildNode(db *DB_CONNECTION, child *Node) {
	var i int
	for i = 0; i < int(parent.numKeys); i++ {
		if string(child.keys[0]) < string(parent.keys[i]) {
			parent.keys = insertIntoKeys(child.keys[0], parent.keys, i)
			parent.childOffsets = insertIntoOffsets(child.childOffsets[0], parent.childOffsets, i)
			parent.childOffsets = insertIntoOffsets(child.childOffsets[1], parent.childOffsets, i+1)
			break
		}
	}
	// insert at rightmost index
	if i == int(parent.numKeys) {
		parent.keys = insertIntoKeys(child.keys[0], parent.keys, i)
		parent.childOffsets = insertIntoOffsets(child.childOffsets[0], parent.childOffsets, i)
		parent.childOffsets = insertIntoOffsets(child.childOffsets[1], parent.childOffsets, i+1)
	}
	parent.numKeys++

	if int32(parent.numKeys) >= Order {
		db.splitNode(parent)
	}
}

func (parent *Node) addLeaf(db *DB_CONNECTION, l *Leaf) {
	var i int
	for i = 0; i < int(parent.numLeaves); i++ {
		if string(l.key) < string(parent.leaves[i].key) {
			parent.leaves = insertIntoLeaves(l, parent.leaves, i)
			break
		}
	}
	// insert at rightmost index
	if i == int(parent.numLeaves) {
		parent.leaves = insertIntoLeaves(l, parent.leaves, i)
	}
	parent.numLeaves++

	if int32(parent.numLeaves) >= Order {
		db.splitLeaves(parent)
	} else {
		db.writeNodeToFile(parent)
	}
}

// Modified from: https://stackoverflow.com/a/61822301
func insertIntoKeys(k []byte, keys [][]byte, i int) [][]byte {
	if len(keys) == i { // nil or empty slice or after last element
		return append(keys, k)
	}
	keys = append(keys[:i+1], keys[i:]...) // index < len(a)
	keys[i] = k
	return keys
}

func insertIntoLeaves(l *Leaf, leaves []*Leaf, i int) []*Leaf {
	if len(leaves) == i {
		return append(leaves, l)
	}
	leaves = append(leaves[:i+1], leaves[i:]...)
	leaves[i] = l
	return leaves
}

func insertIntoOffsets(offset uint32, childOffsets []uint32, i int) []uint32 {
	if len(childOffsets) == i {
		return append(childOffsets, offset)
	}
	childOffsets = append(childOffsets[:i+1], childOffsets[i:]...)
	childOffsets[i] = offset
	return childOffsets
}

func (n *Node) checkHasLeaf() bool {
	return n.numLeaves != 0
}
