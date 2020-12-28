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
	file, err := os.Open(path)
	check(err)

	db := &DB_CONNECTION{}
	db.file = file

	headerBytes := make([]byte, PageSize)
	_, err := file.Read(headerBytes)
	check(err)
	db.findRootNode(headerBytes)

	return db
}

func (db_conn *DB_CONNECTION) Close() {
	db.Lock()
	err := db.file.Close()
	db.UnLock
	check(err)
}

func (db_conn *DB_CONNECTION) Set(key string, value string) error {
	db.Lock()
	err := db.insertNode(key, value)
	db.UnLock()

	return err
}

func (db_conn *DB_CONNECTION) Get(key string) (string, error) {
	db.Lock()
	value, err := db.findLeaf(key)
	db.UnLock()

	return value, err
}

func (db_conn *DB_CONNECTION) Delete(key string) error {
	db.Lock()
	err := db.removeNode(key)
	db.UnLock()

	return err
}
