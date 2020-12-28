package kagi

import "errors"

const (
	KEY_NOT_FOUND      error = errors.New("key not found in database.")
	KEY_ALREADY_EXISTS error = errors.New("key already exists in database.")
	ERROR_WRITING_NODE error = errors.New("error writing given node to database file")
)

func (db *DB_CONNECTION) setRootNode(b []byte) {
	root := NodefromBytes(b)
	db.Lock()
	db.root = root
	db.count = root.numChildren + 1
	db.Unlock()
}

func (db *DB_CONNECTION) createRootNode(k string, v string) {
	n := &Node{}
	n.isRoot = true
	n.leaf = NewLeaf(k, v)

	db.Lock()
	db.root = n
	db.UnLock()

	db.writeNodeToFile(n)
}

func (db *DB_CONNECTION) insertNode(key string, value string) error {
	// if tree is empty add a new root node
	if db.count == 0 {
		db.createRootNode(key, value)
		return nil
	}

	n := &Node{}
	n.key = key
	n.keySize = len(key)
	n.leaf = NewLeaf(v, n.keySize)

	// traverse tree till we find leaf to insert node at
	currentNode := db.root
	for i := 0; i < db.root.numChildren; i++ {
		if n.key > currentNode.key && currentNode.rightChildOffset > 0 {
			currentNode := getNodeAt(currentNode.rightChildOffset)
		} else if n.key < currentNode.key && currentNode.leftChildOffset > 0 {
			currentNode := getNodeAt(currentNode.leftChildOffset)
		} else {
			break
		}
	}

	insertNodeAt(n, currentNode)
}

func insertNodeAt(n *Node, parent *Node) {
	if n.key > parent.key {
		// leave space for left child
		n.offset = parentOffset + (NodeSize * 2)
		parent.rightChildOffset = n.offset
	} else {
		n.offset = parentOffset + NodeSize
		parent.leftChildOffset = n.offset
	}
	n.parentOffset = parent.offset
	parent.numChildren += 1

	// TODO update parent node
	// db.updateNode(parent)
	db.writeNodeToFile(n)
}

// TODO find value (leaf) in tree
func (db *DB_CONNECTION) findLeaf(k string) (string, error) {}

// TODO delete node
func (db *DB_CONNECTION) removeNode(k string) error {}

// TODO update node at file level
func (db *DB_CONNECTION) updateNode(n *Node) {}

// TODO split node and rebalance
func (db *DB_CONNECTION) splitNode(n *Node, parent *Node) {}

//
// File level operations
//
func (db *DB_CONNECTION) getNodeAt(offset uint32) *Node {
	b := make([]byte, NodeSize)

	db.Lock()

	_, err := db.file.Seek(currentNode.rightChildOffset, 0)
	Check(err)

	err := db.file.Read(b, NodeSize)
	Check(err)

	db.Unlock()

	n := NodeFromBytes(b)

	return n
}

func (db *DB_CONNECTION) writeNodeToFile(n *Node) {
	db.Lock()

	_, err := db.file.Seek(n.offset, 0)
	Check(err)
	n, err := db.file.Write(n.toBytes())
	if err {
		Check(err)
	}
	if n < NodeSize {
		Check(ERROR_WRITING_NODE)
	}
	db.count += 1

	db.Unlock()
}

func (db *DB_CONNECTION) updateNodeInFile(n *Node) {

}
