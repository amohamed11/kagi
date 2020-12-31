package kagi

func (db *DB_CONNECTION) setRootNode() {
	db.root = db.getNodeAt(0)
	db.count = int(db.root.numChildren) + 1
}

func (db *DB_CONNECTION) createRootNode(k string, v string) {
	n := &Node{}

	n.isRoot = TRUE

	n.keys = append(n.keys, &Data{data: []byte(k), size: int32(len(k))})
	n.numKeys++

	newLeaf := NewLeaf(k, v)
	n.leaves = append(n.leaves, newLeaf)
	n.numLeaves++

	db.root = n
	db.writeNodeToFile(n)
}

func (db *DB_CONNECTION) insert(k string, v string) error {
	// if tree is empty add a new root node
	if db.count == 0 {
		db.createRootNode(k, v)
		return nil
	}

	newLeaf := NewLeaf(k, v)

	parent := db.searchNode(k, db.root)
	index := 0
	for index = 0; uint16(index) < parent.numLeaves; index++ {
		// found leaf with correct key or no more leaves left
		if k <= string(parent.leaves[index].key.data) {
			break
		}
	}

	if k == string(parent.leaves[index].key.data) {
		return KEY_ALREADY_EXISTS
	}

	db.insertLeaf(newLeaf, parent, index)

	return nil
}

func (db *DB_CONNECTION) insertLeaf(l *Leaf, parent *Node, index int) {
	// TODO handle filled leaf bucket, aka splitting
	// if int32(parent.numLeaves) >= Order {
	// }

	if int32(parent.numLeaves) < Order {
		parent.leaves = insertIntoLeaves(l, parent.leaves, index)

		// update parent & db count
		parent.numLeaves++
		db.writeNodeToFile(parent)
		db.count++
	}
}

func insertIntoLeaves(l *Leaf, leaves []*Leaf, i int) []*Leaf {
	return append(leaves[:i], append([]*Leaf{l}, leaves[i:]...)...)
}

func (db *DB_CONNECTION) findLeaf(k string) (*Leaf, error) {
	parent := db.searchNode(k, db.root)
	index := 0

	for index = 0; uint16(index) < parent.numLeaves; index++ {
		// found leaf with correct key or no more leaves left
		if string(parent.leaves[index].key.data) == k {
			break
		}
	}

	if string(parent.leaves[index].key.data) == k {
		return parent.leaves[index], nil
	}

	return nil, KEY_NOT_FOUND
}

// recursively traverse tree till we find node that has a leaves
func (db *DB_CONNECTION) searchNode(k string, currentNode *Node) *Node {
	if !currentNode.checkHasLeaf() {
		i := 0
		for i = 0; uint16(i) < currentNode.numKeys; i++ {
			if k < string(currentNode.keys[i].data) {
				n := db.getNodeAt(currentNode.childOffsets[i])
				return db.searchNode(k, n)
			}
		}

		// search rightmost child
		n := db.getNodeAt(currentNode.childOffsets[i])
		return db.searchNode(k, n)
	}

	return currentNode
}

func (db *DB_CONNECTION) getNodeAt(offset uint32) *Node {
	b := make([]byte, BlockSize)

	db.readBytesAt(b, offset)

	n := NodeFromBytes(b)

	return n
}

// TODO delete node
//func (db *DB_CONNECTION) removeNode(k string) error {}

// TODO split node and rebalance
//func (db *DB_CONNECTION) splitNode(n *Node, parent *Node) {}

//
// File level operations
//

func (db *DB_CONNECTION) readBytesAt(b []byte, offset uint32) {
	db.Lock()

	_, err1 := db.file.Seek(int64(offset), 0)
	Check(err1)

	_, err2 := db.file.Read(b)
	Check(err2)

	db.Unlock()
}

func (db *DB_CONNECTION) writeNodeToFile(n *Node) {
	db.Lock()

	_, err1 := db.file.Seek(int64(n.offset), 0)
	Check(err1)
	written, err2 := db.file.Write(n.toBytes())
	Check(err2)
	if written < int(BlockSize) {
		Check(ERROR_WRITING_NODE)
	}

	db.Unlock()
}
