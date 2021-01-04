package kagi

import (
	"math/rand"
	"testing"
	"time"
)

const benchPath = "bench.kagi"

var benchOptions = DB_OPTIONS{
	path:  benchPath,
	logs:  "",
	clean: true,
}

func BenchmarkSet1000Keys(b *testing.B) {
	db := Open(benchOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(10000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i += 10 {
			k := seq[i : i+5]
			v := seq[i+5 : i+10]

			err := db.Set(k, v)
			db.logError(err)
		}
		db.Clear()
	}

	db.Close()
}

func BenchmarkGet1000Keys(b *testing.B) {
	db := Open(benchOptions)
	rand.Seed(time.Now().UnixNano())
	seq := randSeq(10000)

	for i := 0; i < 10000; i += 10 {
		k := seq[i : i+5]
		v := seq[i+5 : i+10]

		err := db.Set(k, v)
		db.logError(err)
	}

	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i += 10 {
			k := seq[i : i+5]
			v := seq[i+5 : i+10]

			found, err := db.Get(k)
			if found != v {
				b.Error(err)
				b.Errorf(`actual: "%s", expected: "%s"`, found, v)
			}
		}
	}

	db.Close()
}
