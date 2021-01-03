package kagi

import (
	"os"
	"sync"
)

type DB_CONNECTION struct {
	sync.Mutex
	file  *os.File
	root  *Node
	count uint32
}

func Open(path string) *DB_CONNECTION {
	db := &DB_CONNECTION{}

	file, err1 := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	Check(err1)

	db.file = file
	fileInfo, err2 := db.file.Stat()
	Check(err2)

	if fileInfo.Size() != 0 {
		db.loadDB()
	}

	return db
}

func (db *DB_CONNECTION) Close() {
	// update count
	db.writeBytesAt(BytesFromUint32(db.count), 0)
	err := db.file.Close()
	Check(err)
}

func (db *DB_CONNECTION) Clear() {
	db.Lock()
	err := db.file.Truncate(0)
	Check(err)
	db.count = 0
	db.root = nil
	db.Unlock()
}

func (db *DB_CONNECTION) Set(key string, value string) error {
	err := db.insert(key, value)

	return err
}

func (db *DB_CONNECTION) Get(key string) (string, error) {
	leaf, err := db.findLeaf(key)

	if leaf == nil {
		return "", err
	}
	return string(leaf.value.data), err
}

// func (db *DB_CONNECTION) Delete(key string) error {
// 	db.Lock()
// 	err := db.removeNode(key)
// 	db.Unlock()

// 	return err
// }
