package kagi

import (
	"os"
	"sync"
)

type DB_CONNECTION struct {
	sync.Mutex
	file *os.File
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Open(path string) *DB {
	file, err := os.Open(path)
	check(err)

	db := &DB_CONNECTION{}
	db.file = file

	return db
}

func Close(db *DB) {
	db.Lock()
	err := db.file.Close()
	db.UnLock
	check(err)
}
