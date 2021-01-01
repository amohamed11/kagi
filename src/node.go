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
//  3. Left & Right become child leaf nodes of the Middle
//  4. Add middle node as child to parent
//  5. If parent is now full, split parent node as well
func (db DB_CONNECTION) splitLeaves(parent *Node) {
	half := Order - 1/2
	left := parent.leaves[:half]
	middle := parent.leaves[half]
	right := parent.leaves[half:]

	// create new node using middle key
	middleBranchNode := &Node{
		numKeys:      1,
		numChildren:  parent.numLeaves,
		offset:       uint32(BlockSize) * uint32(db.count),
		parentOffset: parent.offset,
	}
	middleBranchNode.keys = make([]*Data, 1)
	middleBranchNode.keys[0] = middle.key

	// offsets for left & right splits
	middleBranchNode.childOffsets[0] = middleBranchNode.offset + uint32(BlockSize)
	middleBranchNode.childOffsets[1] = middleBranchNode.offset + uint32(BlockSize*2)

	// create left & right nodes & populate with split leaves
	leftLeafNode := &Node{
		numLeaves:    uint16(half),
		offset:       middleBranchNode.childOffsets[0],
		parentOffset: middleBranchNode.offset,
		leaves:       left,
	}

	rightLeafNode := &Node{
		numLeaves:    uint16(Order - half),
		offset:       middleBranchNode.childOffsets[1],
		parentOffset: middleBranchNode.offset,
		leaves:       right,
	}

	parent.addKey(middleBranchNode)

	db.writeNodeToFile(parent)
	db.writeNodeToFile(middleBranchNode)
	db.writeNodeToFile(leftLeafNode)
	db.writeNodeToFile(rightLeafNode)
	db.count += 2
}

// TODO split node & update parent links
// func (n *Node) splitNode() {
// 	half := Order - 1/2
// 	leftKeys := n.keys[:half]
// 	middleKey := n.keys[half]
// 	rightKeys := n.keys[half:]

// 	leftoffsets := n.childOffsets[:half]
// 	middleOffset := n.childOffsets[half]
// 	rightOffsets := n.childOffsets[half:]
// }

func (parent *Node) addKey(n *Node) {
	for i := 0; i < int(parent.numKeys); i++ {
		if string(n.keys[0].data) < string(parent.keys[i].data) {
			parent.keys = insertIntoKeys(n.keys[0], parent.keys, i)
			insertIntoOffsets(n.offset, parent.childOffsets, i)
		}
	}

	// TODO split parent node
	// if int32(n.numKeys) == Order {
	// 	n.splitNode()
	// }
}

func insertIntoKeys(k *Data, keys []*Data, i int) []*Data {
	return append(keys[:i], append([]*Data{k}, keys[i:]...)...)
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

func insertIntoLeaves(l *Leaf, leaves []*Leaf, i int) []*Leaf {
	return append(leaves[:i], append([]*Leaf{l}, leaves[i:]...)...)
}

func (n *Node) checkHasLeaf() bool {
	return n.numLeaves != 0
}
