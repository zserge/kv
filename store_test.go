package kv

import (
	"os"
	"testing"
)

const StoreTestPath = "store-test"

func TestStoreEmpty(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	if keys := store.List(""); len(keys) != 0 {
		t.Error(keys)
	}
	if item := store.Get("foo", &ByteItem{[]byte{}}); item != nil {
		t.Error(nil)
	}
}

func TestStoreSet(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	<-store.Set("foo", &ByteItem{[]byte("Hello")})
	item := store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "Hello" {
		t.Error(item)
	}
	<-store.Set("foo", &ByteItem{[]byte("World")})
	item = store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "World" {
		t.Error(item)
	}
	if keys := store.List(""); len(keys) != 1 || keys[0] != "foo" {
		t.Error(keys)
	}
}

func TestStoreDel(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	<-store.Set("foo", &ByteItem{[]byte("Hello")})
	item := store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "Hello" {
		t.Error(item)
	}
	<-store.Set("foo", nil)
	if item := store.Get("foo", &ByteItem{}); item != nil {
		t.Error(item)
	}
	if err := <-store.Set("missing key", nil); err == nil {
		t.Error("missing key should return error on removal")
	}
	if item := store.Get("foo", &ByteItem{}); item != nil {
		t.Error(item)
	}
	if keys := store.List(""); len(keys) != 0 {
		t.Error(keys)
	}
}

func TestStoreList(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	store.Set("foo", &ByteItem{[]byte("Hello")})
	store.Set("bar", &ByteItem{[]byte("World")})
	store.Set("baz", &ByteItem{[]byte("!")})
	<-store.Flush()
	if keys := store.List(""); len(keys) != 3 {
		t.Error(keys)
	}
	if keys := store.List("foo"); len(keys) != 1 {
		t.Error(keys)
	}
	if keys := store.List("ba"); len(keys) != 2 {
		t.Error(keys)
	}
}

func TestStoreError(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	if err := <-store.Set("foo", &ByteItem{[]byte("Hello")}); err != nil {
		t.Error(err)
	}
	if err := <-store.Set("", &ByteItem{[]byte("Hello")}); err == nil {
		t.Error("Error expected (is a directory), but nil received")
	}
}

func TestStoreFlush(t *testing.T) {
	store := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	store.Set("foo", &ByteItem{[]byte("Hello")})
	store.Set("bar", &ByteItem{[]byte("World")})
	store.Set("baz", &ByteItem{[]byte("!")})
	<-store.Flush()
	<-store.Flush()
	<-store.Flush()
}

func TestLRUWithoutBackend(t *testing.T) {
	store := NewLRU(2, nil)
	store.Set("foo", &ByteItem{[]byte("Hello")})
	store.Set("bar", &ByteItem{[]byte("World")})
	store.Set("baz", &ByteItem{[]byte("!")})
	item := store.Get("baz", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "!" {
		t.Error(item.Value)
	}
	item = store.Get("bar", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "World" {
		t.Error(item.Value)
	}
	// "foo" will be dropped out of Store
	if nilItem := store.Get("foo", &ByteItem{}); nilItem != nil {
		t.Error(nilItem)
	}
	store.Set("foo", &ByteItem{[]byte("Hello")})
	item = store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "Hello" {
		t.Error(item.Value)
	}
	// now "baz" item will be dropped out and it's been least recently updated
	if nilItem := store.Get("baz", &ByteItem{}); nilItem != nil {
		t.Error(nilItem)
	}
	if keys := store.List(""); len(keys) != 2 {
		t.Error(keys)
	}
	if keys := store.List("b"); len(keys) != 1 || keys[0] != "bar" {
		t.Error(keys)
	}
}

func TestLRUWithBackend(t *testing.T) {
	dir := NewStore(StoreTestPath)
	defer os.RemoveAll(StoreTestPath)
	store := NewLRU(2, dir)
	store.Set("foo", &ByteItem{[]byte("Begin")})
	store.Set("foo", &ByteItem{[]byte("Hello")})
	store.Set("bar", &ByteItem{[]byte("World")})

	// This kicks "foo" out of cache to the backend store, updated "foo" value
	// should be written to disk
	<-store.Set("baz", &ByteItem{[]byte("!")})
	if keys := dir.List(""); len(keys) != 1 || keys[0] != "foo" {
		t.Error(keys)
	}

	// check locally cached values
	item := store.Get("baz", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "!" {
		t.Error(item.Value)
	}
	item = store.Get("bar", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "World" {
		t.Error(item.Value)
	}

	// "foo" is not in the cache, but will be read from the backend store anyway
	item = store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "Hello" {
		t.Error(item.Value)
	}

	// when "foo" was read "baz" has been dropped out, so backend store should
	// have 2 items: "foo" and "baz"
	item = dir.Get("foo", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "Hello" {
		t.Error(item.Value)
	}
	item = dir.Get("baz", &ByteItem{}).(*ByteItem)
	if string(item.Value) != "!" {
		t.Error(item.Value)
	}

	if keys := dir.List(""); len(keys) != 2 {
		t.Error(keys)
	}

	// Sync all cached items to disk
	<-store.Flush()
	if keys := dir.List(""); len(keys) != 3 {
		t.Error(keys)
	}
}

func TestItemJSON(t *testing.T) {
	type jsonItem struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}
	a := &jsonItem{"Hello", 1}

	defer os.RemoveAll(StoreTestPath)
	store := NewStore(StoreTestPath)
	<-store.Set("foo", &JSONItem{a})

	byteItem := store.Get("foo", &ByteItem{}).(*ByteItem)
	if string(byteItem.Value) != `{"foo":"Hello","bar":1}`+"\n" {
		t.Error(string(byteItem.Value))
	}
	b := store.Get("foo", &JSONItem{&jsonItem{}}).(*JSONItem).Value.(*jsonItem)
	if b.Foo != "Hello" || b.Bar != 1 {
		t.Error(b)
	}
}

func TestItemGob(t *testing.T) {
	a := []string{"a", "b"}

	defer os.RemoveAll(StoreTestPath)
	store := NewStore(StoreTestPath)
	<-store.Set("foo", &GobItem{a})

	b := []string{}
	store.Get("foo", &GobItem{&b})
	if len(b) != 2 || b[0] != "a" || b[1] != "b" {
		t.Error(b)
	}
}
