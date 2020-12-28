package kagi

import (
	"os"
	"sync"
)

type DB_CONNECTION struct {
	sync.Mutex
	file  *os.File
	root  *Node
	count int
}

func Open(path string) *DB_CONNECTION {
	db := &DB_CONNECTION{}

	file, err1 := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	Check(err1)

	db.file = file
	fileInfo, err2 := db.file.Stat()
	Check(err2)

	if fileInfo.Size() != 0 {
		db.setRootNode()
	}

	return db
}

func (db *DB_CONNECTION) Close() {
	db.Lock()
	err := db.file.Close()
	db.Unlock()
	Check(err)
}

func (db *DB_CONNECTION) Clear() {
	db.Lock()
	err := db.file.Truncate(0)
	db.Unlock()
	Check(err)
}

func (db *DB_CONNECTION) Set(key string, value string) error {
	err := db.insertNode(key, value)

	return err
}

func (db *DB_CONNECTION) Get(key string) (string, error) {
	node, err := db.findLeaf(key)

	if node == nil {
		return "", err
	}
	return node.leaf.value, err
}

// func (db *DB_CONNECTION) Delete(key string) error {
// 	db.Lock()
// 	err := db.removeNode(key)
// 	db.Unlock()

// 	return err
// }
