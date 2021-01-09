package kagi

// count is saved at 0, and root right after it
func (db *DB_CONNECTION) loadDB() {
	db.root = db.getNodeAt(0)
	db.count = db.root.dbCount

	db.logInfo("loaded root node")
}

func (db *DB_CONNECTION) createRootNode(k string, v string) {
	n := &Node{}

	n.isRoot = TRUE

	n.leaves = make([]*Leaf, 1)
	n.leaves[0] = NewLeaf(k, v)
	n.numLeaves++

	n.dbCount = 1
	n.offset = 0

	db.root = n
	db.count = 1
	db.writeNodeToFile(n)
}

func (db *DB_CONNECTION) insert(k string, v string) error {
	// if tree is empty add a new root node
	db.logInfo("inserting (key: %s, value: %s)\n", k, v)

	if db.count == 0 {
		db.logInfo("creating a root node")
		db.createRootNode(k, v)
		return nil
	}

	newLeaf := NewLeaf(k, v)
	parent := db.searchNode(k, db.root, true)
	var keyFound string

	for i := 0; i < int(parent.numLeaves); i++ {
		// found leaf with correct key or no more leaves left
		if k <= string(parent.leaves[i].key) {
			keyFound = string(parent.leaves[i].key)
			break
		}
	}

	if k == keyFound {
		return KEY_ALREADY_EXISTS
	}

	parent.addLeaf(db, newLeaf)

	return nil
}

// Find node that contain this key as leaf
// Remove the leaf from the node
func (db *DB_CONNECTION) remove(k string) error {
	var index int
	newKey := make([]byte, MaxKeySize)
	neigbhour := &Node{}
	leafNode := db.searchNode(k, db.root, true)

	for index = 0; index < int(leafNode.numLeaves); index++ {
		if string(leafNode.leaves[index].key) == k {
			leafNode.leaves = append(leafNode.leaves[:index], leafNode.leaves[index+1:]...)
			leafNode.numLeaves--
			db.writeNodeToFile(leafNode)
			break
		}
	}

	// key was found
	if index == int(leafNode.numLeaves) {
		return KEY_NOT_FOUND
	}

	// handle underflow
	if leafNode.isRoot == FALSE && int32(leafNode.numLeaves) < Degree {
		var i int
		parent := db.getNodeAt(leafNode.parentOffset)

		for i = 0; i < len(parent.childOffsets); i++ {
			if leafNode.offset == parent.childOffsets[i] {
				break
			}
		}

		if i+1 < len(parent.childOffsets) {
			neigbhour = db.getNodeAt(parent.childOffsets[i+1])
		} else {
			neigbhour = db.getNodeAt(parent.childOffsets[i-1])
		}

		neigbhour.leaves = combineLeaves(neigbhour.leaves, leafNode.leaves)
		neigbhour.numLeaves = uint16(len(neigbhour.leaves))
		newKey = neigbhour.leaves[0].key

		if int32(neigbhour.numLeaves) >= Order {
			db.splitLeaves(neigbhour)
		} else {
			db.writeNodeToFile(neigbhour)
		}
	}

	// handle branching key deletion & rebalancing
	if leafNode.isRoot == FALSE && index == 0 {
		db.replaceBranchingKey(k, string(newKey))
	}

	return nil
}

func (db *DB_CONNECTION) findLeaf(k string) (*Leaf, error) {
	parent := db.searchNode(k, db.root, true)

	for index := 0; index < int(parent.numLeaves); index++ {
		// found leaf with correct key or no more leaves left
		if string(parent.leaves[index].key) == k {
			return parent.leaves[index], nil
		}
	}

	return nil, KEY_NOT_FOUND
}

// recursively traverse tree till we find leaf node with given key
// if isLeaf is false, will search for a branching node that has the given key
func (db *DB_CONNECTION) searchNode(k string, currentNode *Node, isLeaf bool) *Node {
	if !currentNode.checkHasLeaf() {
		i := 0
		for i = 0; i < int(currentNode.numKeys); i++ {
			if !isLeaf {
				// searching for a branching node with key k
				if k == string(currentNode.keys[i]) {
					return currentNode
				} else if k == string(currentNode.keys[i]) {
					n := db.getNodeAt(currentNode.childOffsets[i])
					return db.searchNode(k, n, isLeaf)
				}

			} else if k < string(currentNode.keys[i]) {
				n := db.getNodeAt(currentNode.childOffsets[i])
				return db.searchNode(k, n, isLeaf)
			}
		}

		// search rightmost child
		n := db.getNodeAt(currentNode.childOffsets[i])
		return db.searchNode(k, n, isLeaf)
	}

	return currentNode
}

func (db *DB_CONNECTION) replaceBranchingKey(oldKey string, newKey string) {
	n := db.searchNode(oldKey, db.root, false)

	for i := 0; i < len(n.keys); i++ {
		if string(n.keys[i]) == oldKey {
			n.keys[i] = []byte(newKey)
			db.writeNodeToFile(n)
			break
		}
	}
}

func (db *DB_CONNECTION) getNodeAt(offset uint32) *Node {
	b := make([]byte, BlockSize)
	db.readBytesAt(b, int64(offset))
	n := NodeFromBytes(b)

	return n
}

func (db *DB_CONNECTION) writeNodeToFile(n *Node) {
	nodeBytes := n.toBytes()
	if n.offset >= db.count*uint32(BlockSize) {
		db.count++
	}
	db.writeBytesAt(nodeBytes, int64(n.offset))
}

func (db *DB_CONNECTION) readBytesAt(b []byte, offset int64) {
	db.logInfo("reading bytes at offset: %d\n", offset)

	_, err1 := db.file.Seek(offset, 0)
	db.logError(err1)

	_, err2 := db.file.Read(b)
	db.logError(err2)
}

func (db *DB_CONNECTION) writeBytesAt(b []byte, offset int64) {
	db.logInfo("writing bytes at offset: %d\n", offset)

	_, err1 := db.file.Seek(offset, 0)
	db.logError(err1)

	written, err2 := db.file.Write(b)
	db.logError(err2)

	if written < len(b) {
		db.logError(ERROR_WRITING)
	}
}
