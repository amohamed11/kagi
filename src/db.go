package kagi

import (
	"os"
	"sync"
)

type DB_CONNECTION struct {
	sync.Mutex
	file     *os.File
	filePath string
	root     *Node
	count    uint32
}

func Open(path string) *DB_CONNECTION {
	db := &DB_CONNECTION{}

	file, err1 := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	Check(err1)

	db.filePath = path
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
	defer db.file.Close()
}

func (db *DB_CONNECTION) Clear() {
	db.Lock()

	db.count = 0
	db.root = nil

	err := db.file.Truncate(0)
	Check(err)

	db.Unlock()
}

func (db *DB_CONNECTION) Set(key string, value string) error {
	db.Lock()
	err := db.insert(key, value)
	db.Unlock()

	return err
}

func (db *DB_CONNECTION) Get(key string) (string, error) {
	db.Lock()
	leaf, err := db.findLeaf(key)
	db.Unlock()

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
