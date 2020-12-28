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
		headerBytes := make([]byte, NodeSize)
		_, err3 := file.Read(headerBytes)
		Check(err3)
		db.setRootNode(headerBytes)
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
	db.Lock()
	err := db.insertNode(key, value)
	db.Unlock()

	return err
}

func (db *DB_CONNECTION) Get(key string) (string, error) {
	db.Lock()
	node, err := db.findLeaf(key)
	db.Unlock()

	return node.leaf.value, err
}

// func (db *DB_CONNECTION) Delete(key string) error {
// 	db.Lock()
// 	err := db.removeNode(key)
// 	db.Unlock()

// 	return err
// }
