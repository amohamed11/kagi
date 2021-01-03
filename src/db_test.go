package kagi

import (
	"math/rand"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const testPath = "test.kagi"

// thanks to: https://stackoverflow.com/a/22892986
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestSet1Key(t *testing.T) {
	db := Open(testPath)
	db.Clear()
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(10)
	k := seq[0:5]
	v := seq[5:]

	err := db.Set(k, v)
	Check(err)
	db.Close()
}

func TestGet1Key(t *testing.T) {
	db := Open(testPath)
	db.Clear()
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(10)
	k := seq[0:5]
	v := seq[5:]

	err := db.Set(k, v)
	Check(err)

	found, err := db.Get(k)
	if found != v {
		t.Error(err)
		t.Errorf(`actual: "%s", expected: "%s"`, found, v)
	}
	db.Close()
}

func TestSet10Keys(t *testing.T) {
	db := Open(testPath)
	db.Clear()
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(100)

	for i := 0; i < 100; i += 10 {
		k := seq[i : i+5]
		v := seq[i+5 : i+10]

		err := db.Set(k, v)
		Check(err)
	}
	db.Close()
}

func TestGet10Keys(t *testing.T) {
	db := Open(testPath)
	db.Clear()
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(100)

	for i := 0; i < 100; i += 10 {
		k := seq[i : i+5]
		v := seq[i+5 : i+10]

		err := db.Set(k, v)
		Check(err)

		found, err := db.Get(k)
		if found != v {
			t.Error(err)
			t.Errorf(`actual: "%s", expected: "%s"`, found, v)
		}
	}
	db.Close()
}
