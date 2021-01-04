package kagi

import "fmt"

type Error string

func (e Error) Error() string { return string(e) }

const (
	TRUE               = 1
	FALSE              = 0
	KEY_NOT_FOUND      = Error("key not found in database.")
	KEY_ALREADY_EXISTS = Error("key already exists in database.")
	ERROR_WRITING      = Error("error writing to database")
)

func (db *DB_CONNECTION) logError(e error) {
	if e != nil {
		if db.errorLogger != nil {
			db.errorLogger.Fatal(e.Error())
		} else {
			panic(e)
		}
	}
}

func (db *DB_CONNECTION) logInfo(info string, v ...interface{}) {
	if db.infoLogger != nil {
		s := fmt.Sprintf(info, v...)
		db.infoLogger.Println(s)
	}
}
