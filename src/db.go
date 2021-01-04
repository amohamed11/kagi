package kagi

import (
	"log"
	"os"
	"sync"
)

type DB_CONNECTION struct {
	sync.Mutex
	file        *os.File
	filePath    string
	root        *Node
	count       uint32
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

type DB_OPTIONS struct {
	path  string // path for database, otherwise a default is chosen
	logs  string // path to logs file, no logging if left empty
	clean bool   // clean database
}

func Open(options DB_OPTIONS) *DB_CONNECTION {
	db := &DB_CONNECTION{}

	if options.path != "" {
		db.filePath = options.path
	}

	if options.logs != "" {
		logFile, err := os.OpenFile(options.logs, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err.Error())
		} else {
			db.infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
			db.errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	file, err1 := os.OpenFile(db.filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	db.logError(err1)

	db.file = file

	fileInfo, err2 := db.file.Stat()
	db.logError(err2)

	if options.clean {
		db.Clear()
	} else if fileInfo.Size() != 0 {
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
	db.logError(err)

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
