package kv

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

const StoreBenchPath = "store-bench"

func BenchmarkStoreGet(b *testing.B) {
	store := NewStore(StoreBenchPath)
	item := &ByteItem{bytes.Repeat([]byte("hello"), 1000)}
	store.Set("foo", item)

	b.SetBytes(int64(len(item.Value)))
	for i := 0; i < b.N; i++ {
		store.Get("foo", item)
	}
}

func BenchmarkStoreSet(b *testing.B) {
	store := NewStore(StoreBenchPath)
	item := &ByteItem{bytes.Repeat([]byte("hello"), 1000)}

	b.SetBytes(int64(len(item.Value)))
	for i := 0; i < b.N; i++ {
		store.Set("foo", item)
	}
}

func BenchmarkStoreCacheGet(b *testing.B) {
	store := NewLRU(1, NewStore(StoreBenchPath))
	item := &ByteItem{bytes.Repeat([]byte("hello"), 1000)}
	store.Set("foo", item)

	b.SetBytes(int64(len(item.Value)))
	for i := 0; i < b.N; i++ {
		store.Get("foo", item)
	}
}

func BenchmarkStoreCacheSet(b *testing.B) {
	store := NewLRU(1, NewStore(StoreBenchPath))
	item := &ByteItem{bytes.Repeat([]byte("hello"), 1000)}

	b.SetBytes(int64(len(item.Value)))
	for i := 0; i < b.N; i++ {
		store.Set("foo", item)
	}
}

func BenchmarkWrite__32B(b *testing.B) {
	benchWrite(b, 32, true)
}

func BenchmarkWrite__1KB(b *testing.B) {
	benchWrite(b, 1024, true)
}

func BenchmarkWrite__4KB(b *testing.B) {
	benchWrite(b, 4096, true)
}

func BenchmarkWrite_10KB(b *testing.B) {
	benchWrite(b, 10240, true)
}

func BenchmarkRead__32B_NoCache(b *testing.B) {
	benchRead(b, 32, 1)
}

func BenchmarkRead__1KB_NoCache(b *testing.B) {
	benchRead(b, 1024, 1)
}

func BenchmarkRead__4KB_NoCache(b *testing.B) {
	benchRead(b, 4096, 1)
}

func BenchmarkRead_10KB_NoCache(b *testing.B) {
	benchRead(b, 10240, 1)
}

func BenchmarkRead__32B_WithCache(b *testing.B) {
	benchRead(b, 32, keyCount)
}

func BenchmarkRead__1KB_WithCache(b *testing.B) {
	benchRead(b, 1024, keyCount)
}

func BenchmarkRead__4KB_WithCache(b *testing.B) {
	benchRead(b, 4096, keyCount)
}

func BenchmarkRead_10KB_WithCache(b *testing.B) {
	benchRead(b, 10240, keyCount)
}

func shuffle(keys []string) {
	ints := rand.Perm(len(keys))
	for i := range keys {
		keys[i], keys[ints[i]] = keys[ints[i]], keys[i]
	}
}

func genValue(size int) []byte {
	v := make([]byte, size)
	for i := 0; i < size; i++ {
		v[i] = uint8((rand.Int() % 26) + 97) // a-z
	}
	return v
}

const (
	keyCount = 1000
)

func genKeys() []string {
	keys := make([]string, keyCount)
	for i := 0; i < keyCount; i++ {
		keys[i] = fmt.Sprintf("%d", i)
	}
	return keys
}

func load(store Store, keys []string, val []byte) {
	for _, key := range keys {
		store.Set(key, &ByteItem{val})
	}
}

func benchRead(b *testing.B, size, cachesz int) {
	b.StopTimer()
	store := NewLRU(cachesz, NewStore(StoreBenchPath))
	keys := genKeys()
	value := genValue(size)
	load(store, keys, value)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		store.Get(keys[i%len(keys)], &ByteItem{})
	}
	b.StopTimer()
}

func benchWrite(b *testing.B, size int, withIndex bool) {
	b.StopTimer()

	store := NewStore(StoreBenchPath)
	keys := genKeys()
	value := genValue(size)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		store.Set(keys[i%len(keys)], &ByteItem{value})
	}
	b.StopTimer()
}
