package kagi

const (
	Order      int32 = 3                  // the upper limit of children for nodes, 2-Order max children
	BlockSize  int32 = 4096               // max size of a node
	Int32Size  int32 = 4                  // size of uint32 used for offsets in node
	Int16Size  int32 = 2                  // size of uint16 used for flags and counts in nodes
	HeaderSize int32 = 4078 - (Order * 4) // 4096 - flags(2*2) - counts(3*2) - offsets(2*4) - childOffsets(Order*4)
)

type Node struct {
	//----Header----
	// Flags
	isRoot    uint16
	isDeleted uint16

	// Counts
	numKeys     uint16
	numChildren uint16
	numLeaves   uint16

	// Offsets
	offset       uint32
	parentOffset uint32
	childOffsets [Order]uint32
	// -------------

	// branching node
	keys []*Data

	// leaf nodes
	leaves []*Leaf
}

type Leaf struct {
	// data
	key   *Data
	value *Data

	// TODO Calculate in memory
	// freeSpace int32
}

type Data struct {
	size int32
	data []byte
}

func NewLeaf(k string, v string) *Leaf {
	l := &Leaf{}

	l.key = &Data{
		data: []byte(k),
		size: int32(len(k)),
	}

	l.value = &Data{
		data: []byte(v),
		size: int32(len(v)),
	}

	return l
}

// Steps for splitting:
//  1. Splits leaves into 3 parts
//  2. Middle becomes a branching node
//  3. Left & Right become child nodes of the Middle
//  4. Add middle node as child to parent
//  5. If parent is now full, split parent node as well
func (db *DB_CONNECTION) splitNode(fullNode *Node) {
	half := (Order - 1) / 2
	leftKey := fullNode.keys[:half]
	middleKey := fullNode.keys[half]
	rightKey := fullNode.keys[half:]

	// create new node using middle key
	middleBranchNode := &Node{
		numKeys:      1,
		numChildren:  2,
		offset:       fullNode.offset,
		parentOffset: fullNode.parentOffset,
	}
	middleBranchNode.keys = make([]*Data, 1)
	middleBranchNode.keys[0] = middleKey

	// offsets for children are at end of the file
	childOffset := uint32(BlockSize) * uint32(db.count)
	middleBranchNode.childOffsets[0] = childOffset
	middleBranchNode.childOffsets[1] = childOffset + uint32(BlockSize)

	// create left & right nodes & link old node's children
	leftChildNode := &Node{
		numKeys:      uint16(half),
		numChildren:  uint16(half + 1),
		offset:       middleBranchNode.childOffsets[0],
		parentOffset: middleBranchNode.offset,
		keys:         leftKey,
	}
	copy(leftChildNode.childOffsets[0:], fullNode.childOffsets[:half+1])

	rightChildNode := &Node{
		numKeys:      uint16(Order - half),
		numChildren:  uint16(Order - half + 1),
		offset:       middleBranchNode.childOffsets[1],
		parentOffset: middleBranchNode.offset,
		keys:         rightKey,
	}
	copy(rightChildNode.childOffsets[0:], fullNode.childOffsets[half+1:])

	// update parent with the middle key
	parent := db.getNodeAt(middleBranchNode.parentOffset)
	parent.addChildNode(db, middleBranchNode)
	db.writeNodeToFile(parent)

	// write newly create nodes
	db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftChildNode)
	db.writeNodeToFile(rightChildNode)
	db.count += 2
}

func (db *DB_CONNECTION) splitLeaves(parent *Node) {
	half := (Order - 1) / 2
	leftLeaf := parent.leaves[:half]
	middleLeaf := parent.leaves[half]
	rightLeaf := parent.leaves[half:]

	// create new node using middle key
	middleBranchNode := &Node{
		numKeys:      1,
		numChildren:  2,
		offset:       uint32(BlockSize) * uint32(db.count),
		parentOffset: parent.offset,
	}
	middleBranchNode.keys = make([]*Data, 1)
	middleBranchNode.keys[0] = middleLeaf.key

	// offsets for left & right splits under the middle
	middleBranchNode.childOffsets[0] = middleBranchNode.offset + uint32(BlockSize)
	middleBranchNode.childOffsets[1] = middleBranchNode.offset + uint32(BlockSize*2)

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
	parent.addChildNode(db, middleBranchNode)
	db.writeNodeToFile(parent)

	// write newly create nodes
	db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftLeafNode)
	db.writeNodeToFile(rightLeafNode)
	db.count += 2
}

// child node should have a single key
func (parent *Node) addChildNode(db *DB_CONNECTION, child *Node) {
	for i := 0; i < int(parent.numKeys); i++ {
		if string(child.keys[0].data) < string(parent.keys[i].data) {
			parent.keys = insertIntoKeys(child.keys[0], parent.keys, i)
			insertIntoOffsets(child.offset, parent.childOffsets, i)
			break
		}
	}
	parent.numKeys++

	if int32(parent.numKeys) == Order {
		db.splitNode(parent)
	}
}

func (parent *Node) addLeaf(db *DB_CONNECTION, l *Leaf) {
	for i := 0; i < int(parent.numLeaves); i++ {
		if string(l.key.data) < string(parent.leaves[i].key.data) {
			parent.leaves = insertIntoLeaves(l, parent.leaves, i)
			break
		}
	}
	parent.numLeaves++

	if int32(parent.numLeaves) == Order {
		db.splitLeaves(parent)
	}
}

func insertIntoKeys(k *Data, keys []*Data, i int) []*Data {
	return append(keys[:i], append([]*Data{k}, keys[i:]...)...)
}

func insertIntoLeaves(l *Leaf, leaves []*Leaf, i int) []*Leaf {
	return append(leaves[:i], append([]*Leaf{l}, leaves[i:]...)...)
}

func insertIntoOffsets(offset uint32, childOffsets [Order]uint32, index int) {
	if childOffsets[index] == uint32(0) {
		childOffsets[index] = offset
	} else {
		tmp := childOffsets[index]
		childOffsets[index] = offset
		insertIntoOffsets(tmp, childOffsets, index+1)
	}
}

func (n *Node) checkHasLeaf() bool {
	return n.numLeaves != 0
}
