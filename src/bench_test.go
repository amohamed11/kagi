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
	clean: false,
}

var seq string

func BenchmarkSet1000Keys(b *testing.B) {
	db := Open(benchOptions)

	for n := 0; n < b.N; n++ {
		rand.Seed(time.Now().UnixNano())
		seq = randSeq(10000)
		db.Clear()
		for i := 0; i < 10000; i += 10 {
			k := seq[i : i+5]
			v := seq[i+5 : i+10]

			err := db.Set(k, v)
			db.logError(err)
		}
	}

	db.Close()
}

func BenchmarkGet1000Keys(b *testing.B) {
	db := Open(benchOptions)

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
