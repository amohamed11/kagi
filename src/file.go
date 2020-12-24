package kagi

import (
	"io/ioutil"
	"sync"
)

type DB_FILE struct {
	dblock    sync.Mutex
	path      string
	dataBytes []byte
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Open(path string) (*DB_FILE, error) {
	compressedData, err := ioutil.ReadFile(path)
	check(err)

	data, err := Decompress(nil, compressedData)
	check(err)
}

func Close()
