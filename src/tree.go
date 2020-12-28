package kagi

import "fmt"

func (db *DB_CONNECTION) setRootNode() {
	db.root = db.getNodeAt(0)
	db.count = int(db.root.numChildren) + 1
}

func (db *DB_CONNECTION) createRootNode(k string, v string) {
	n := &Node{}
	n.isRoot = TRUE
	n.key = k
	n.keySize = int32(len(k))
	n.leaf = NewLeaf(v, n.keySize)

	db.root = n
	db.writeNodeToFile(n)
}

func (db *DB_CONNECTION) insertNode(k string, v string) error {
	// if tree is empty add a new root node
	if db.count == 0 {
		db.createRootNode(k, v)
		return nil
	}

	n := &Node{}
	n.key = k
	n.keySize = int32(len(k))
	n.leaf = NewLeaf(v, n.keySize)

	leafNode := db.searchNode(n.key, db.root)

	if n.key == leafNode.key {
		return KEY_ALREADY_EXISTS
	}

	db.insertNodeAt(n, leafNode)
	return nil
}

func (db *DB_CONNECTION) insertNodeAt(n *Node, parent *Node) {
	if n.key > parent.key {
		// leave space for left child
		n.offset = parent.offset + (uint32(NodeSize) * 2)
		parent.rightChildOffset = n.offset
	} else {
		n.offset = parent.offset + uint32(NodeSize)
		parent.leftChildOffset = n.offset
	}
	n.parentOffset = parent.offset
	parent.numChildren += uint32(1)

	// update parent node
	// fmt.Printf("key: %s, offset: %d, parentOffset: %d\n", n.key, n.offset, parent.offset)
	db.writeNodeToFile(parent)
	db.writeNodeToFile(n)
	db.count += 1
}

func (db *DB_CONNECTION) findLeaf(k string) (*Node, error) {
	leafNode := db.searchNode(k, db.root)

	if leafNode.key == k {
		return leafNode, nil
	}

	return nil, KEY_NOT_FOUND
}

// recursively traverse tree till we find leaf
func (db *DB_CONNECTION) searchNode(k string, currentNode *Node) *Node {
	// fmt.Printf("found: %s, wanted: %s\n", currentNode.key, k)
	if currentNode.numChildren == 0 {
		return currentNode
	}

	nextNode := &Node{}
	fmt.Printf("right: %d, left: %d\n", currentNode.rightChildOffset, currentNode.leftChildOffset)
	if k > currentNode.key && currentNode.rightChildOffset > 0 {
		nextNode = db.getNodeAt(currentNode.rightChildOffset)
	} else if k < currentNode.key && currentNode.leftChildOffset > 0 {
		nextNode = db.getNodeAt(currentNode.leftChildOffset)
	} else {
		return currentNode
	}

	return db.searchNode(k, nextNode)
}

// TODO delete node
//func (db *DB_CONNECTION) removeNode(k string) error {}

// TODO split node and rebalance
//func (db *DB_CONNECTION) splitNode(n *Node, parent *Node) {}

//
// File level operations
//
func (db *DB_CONNECTION) getNodeAt(offset uint32) *Node {
	b := make([]byte, NodeSize)

	db.Lock()

	_, err1 := db.file.Seek(int64(offset), 0)
	Check(err1)

	_, err2 := db.file.Read(b)
	Check(err2)

	db.Unlock()

	n := NodeFromBytes(b)

	return n
}

func (db *DB_CONNECTION) writeNodeToFile(n *Node) {
	db.Lock()

	_, err1 := db.file.Seek(int64(n.offset), 0)
	Check(err1)
	written, err2 := db.file.Write(n.toBytes())
	Check(err2)
	if written < int(NodeSize) {
		Check(ERROR_WRITING_NODE)
	}

	db.Unlock()
}
