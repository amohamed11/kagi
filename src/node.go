package kagi

import (
	"log"
)

const (
	Order         int32 = 5                  // the upper limit of children for nodes, 2-Order max children
	BlockSize     int32 = 4096               // max size of a node
	Int32Size     int32 = 4                  // size of uint32 used for offsets in node
	Int16Size     int32 = 2                  // size of uint16 used for flags and counts in nodes
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
	size uint32
	data []byte
}

func NewLeaf(k string, v string) *Leaf {
	l := &Leaf{}

	l.key = &Data{
		data: []byte(k),
		size: uint32(len(k)),
	}

	l.value = &Data{
		data: []byte(v),
		size: uint32(len(v)),
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
	half := int32(Order / 2)
	leftKey := fullNode.keys[:half]
	middleKey := fullNode.keys[half]
	rightKey := fullNode.keys[half:]

	log.Println("splitting branching node")
	log.Printf("parent node now is key: %s\n", middleKey.data)

	// create new node using middle key
	middleBranchNode := &Node{
		numKeys:      1,
		offset:       fullNode.offset,
		parentOffset: fullNode.parentOffset,
	}
	middleBranchNode.keys = make([]*Data, 1)
	middleBranchNode.keys[0] = middleKey

	// offsets for children are at end of the file
	middleBranchNode.childOffsets = make([]uint32, 2)
	middleBranchNode.childOffsets[0] = uint32(BlockSize)*db.count + uint32(Int32Size)
	middleBranchNode.childOffsets[1] = middleBranchNode.childOffsets[0] + uint32(BlockSize)

	// create left & right nodes & link old node's children
	leftChildNode := &Node{
		numKeys:      uint16(half),
		offset:       middleBranchNode.childOffsets[0],
		parentOffset: middleBranchNode.offset,
		keys:         leftKey,
	}
	leftChildNode.childOffsets = make([]uint32, 0, leftChildNode.numKeys+1)
	leftChildNode.childOffsets = append(leftChildNode.childOffsets, fullNode.childOffsets[:half]...)

	rightChildNode := &Node{
		numKeys:      uint16(Order - half),
		offset:       middleBranchNode.childOffsets[1],
		parentOffset: middleBranchNode.offset,
		keys:         rightKey,
	}
	rightChildNode.childOffsets = make([]uint32, 0, rightChildNode.numKeys+1)
	rightChildNode.childOffsets = append(rightChildNode.childOffsets, fullNode.childOffsets[half:]...)

	if fullNode.isRoot == TRUE {
		middleBranchNode.isRoot = TRUE
	} else {
		// update parent with the middle key
		parent := db.getNodeAt(middleBranchNode.parentOffset)
		parent.addChildNode(db, middleBranchNode)
		db.writeNodeToFile(parent)
	}

	// write newly create nodes
	db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftChildNode)
	db.writeNodeToFile(rightChildNode)
	db.count += 2
}

func (db *DB_CONNECTION) splitLeaves(fullLeafNode *Node) {
	half := int32(Order / 2)
	leftLeaf := fullLeafNode.leaves[:half]
	middleLeaf := fullLeafNode.leaves[half]
	rightLeaf := fullLeafNode.leaves[half:]

	log.Println("splitting leaves")
	log.Printf("creating new branching node with key: %s\n", middleLeaf.key.data)

	// create new node using middle key
	middleBranchNode := &Node{
		numKeys:      1,
		offset:       fullLeafNode.offset,
		parentOffset: fullLeafNode.parentOffset,
	}
	middleBranchNode.keys = make([]*Data, 1)
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

	if fullLeafNode.isRoot == TRUE {
		middleBranchNode.isRoot = TRUE
	} else {
		// update parent with the middle key
		fullLeafNode.addChildNode(db, middleBranchNode)
		db.writeNodeToFile(fullLeafNode)
	}

	// write newly create nodes
	db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftLeafNode)
	db.writeNodeToFile(rightLeafNode)
	db.count += 1
}

// child node should have a single key
func (parent *Node) addChildNode(db *DB_CONNECTION, child *Node) {
	var i int
	for i = 0; i < int(parent.numKeys); i++ {
		if string(child.keys[0].data) < string(parent.keys[i].data) {
			parent.keys = insertIntoKeys(child.keys[0], parent.keys, i)
			parent.childOffsets = insertIntoOffsets(child.childOffsets[0], parent.childOffsets, i)
			parent.childOffsets = insertIntoOffsets(child.childOffsets[1], parent.childOffsets, i+1)
			break
		}
	}
	// insert at rightmost index
	if i == int(parent.numKeys) {
		parent.keys = insertIntoKeys(child.keys[0], parent.keys, i)
		parent.childOffsets = insertIntoOffsets(child.offset, parent.childOffsets, i)
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
		if string(l.key.data) < string(parent.leaves[i].key.data) {
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
func insertIntoKeys(k *Data, keys []*Data, i int) []*Data {
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
