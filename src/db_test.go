package kagi

import (
	"math/rand"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const testPath = "test.kagi"

var testOptions = DB_OPTIONS{
	path:  testPath,
	logs:  "test_logs.txt",
	clean: false,
}

var testClearOptions = DB_OPTIONS{
	path:  testPath,
	logs:  "test_logs.txt",
	clean: true,
}

// thanks to: https://stackoverflow.com/a/22892986
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestSet100Keys(t *testing.T) {
	db := Open(testClearOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(19200)

	// Set keys to be getted
	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		err := db.Set(k, v)
		db.logError(err)
	}
	db.Close()
}

func TestGet100Keys(t *testing.T) {
	db := Open(testClearOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(19200)

	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		err := db.Set(k, v)
		db.logError(err)
	}

	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		found, err2 := db.Get(k)
		if found != v {
			t.Error(err2)
			t.Errorf(`test %d, actual: "%s", expected: "%s"`, i/10, found, v)
		}
	}
	db.Close()
}

func TestGet100KeysAfterClosing(t *testing.T) {
	db := Open(testClearOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(19200)

	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		err := db.Set(k, v)
		db.logError(err)
	}

	db.Close()
	db = Open(testOptions)

	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		found, err2 := db.Get(k)
		if found != v {
			t.Error(err2)
			t.Errorf(`actual: "%s", expected: "%s"`, found, v)
		}
	}
	db.Close()
}

func TestDelete100Keys(t *testing.T) {
	db := Open(testClearOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(19200)

	// Set keys to be deleted
	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]
		v := seq[i+44 : i+144]

		err := db.Set(k, v)
		db.logError(err)
	}

	for i := 0; i < 100; i += 192 {
		k := seq[i : i+44]

		err1 := db.Delete(k)
		if err1 != nil {
			t.Error(err1)
		}

		found, err2 := db.Get(k)
		if found != "" {
			if err2 != KEY_NOT_FOUND {
				t.Error(err2)
			}
			t.Errorf(`actual: "%s", expected: ""`, found)
		}
	}
	db.Close()
}
