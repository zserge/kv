package kv

import (
	"container/list"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Item is something that can be put into a Store. Items should be able to
// read their values from io.Reader and write them into io.Writer.
// Several helper implemenations are provided - for raw bytes, for JSON and for
// gob format.
type Item interface {
	io.ReaderFrom
	io.WriterTo
}

//
// Store is the interface that wraps basic get/set functions for a simple
// key-value storage.
//
// Store can be implemented as on-disk persistent storage, or as in-memory
// cache.
//
// Get fulfils the item with the value associated with the key and returns the
// item. Caller must provide the item instance beforehand, because the way how
// item serializes/deserializes itself depends on its type. Get returns no
// errors explicitly, but nil is returned if key is missing or any other I/O
// error happended. Get is synchronous, your goroutine is blocked until item is
// fully read from the store.
//
// Set writes item contents to the store at the given key. It's an asynchronous
// operation, but caller can read from the returned channel to wait for write
// completion and to get notified if I/O error occurred. If item is nil the key
// is removed from the store
//
// List returns list of keys that exists and in the store and start with the
// given prefix. If prefix is an empty string - all keys are returned. List
// function is syncrhonous.
//
// Flush waits for all writing goroutines to finish and syncs all store data to
// the disk. Flush can be called asynchronously, or the caller can wait for the
// actual flush to happen by reading from the returned channel.
type Store interface {
	Get(key string, item Item) Item
	Set(key string, item Item) <-chan error
	List(prefix string) []string
	Flush() <-chan error
}

// Store implementation that keeps each item in its own file in the given
// directory
type dirStore struct {
	mutex sync.RWMutex
	wg    sync.WaitGroup
	path  string
}

// Creates a new store from the given path. Keys are file names relative to the
// path, values are file contents.
func NewStore(path string) Store {
	return &dirStore{
		path: path,
	}
}

func mkpath(root, s string) string {
	parts := strings.Split(s, "/")
	for i, s := range parts {
		parts[i] = url.QueryEscape(s)
	}
	return filepath.Join(root, filepath.Join(parts...))
}

func (store *dirStore) Get(key string, item Item) Item {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if f, err := os.Open(mkpath(store.path, key)); err == nil {
		defer f.Close()
		if _, err := item.ReadFrom(f); err == nil {
			return item
		}
	}
	return nil
}

func (store *dirStore) Set(key string, item Item) <-chan error {
	c := make(chan error, 1)

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.wg.Add(1)
	go func() {
		defer store.wg.Done()
		defer close(c)
		s := mkpath(store.path, key)
		if item == nil {
			if err := os.Remove(s); err != nil {
				c <- err
			}
		} else {
			os.MkdirAll(filepath.Dir(s), 0700)

			if f, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE, 0600); err != nil {
				c <- err
			} else {
				// FIXME make this atomic (using file move/rename)
				defer f.Close()
				item.WriteTo(f)
			}
		}
	}()
	return c
}

func (store *dirStore) Flush() <-chan error {
	// In disk store flush does not really return any errors because it can not
	// know which write goroutines are now running so it can't collect their
	// errors, also Sync() doesn't return errors either. So error channel is just
	// to meet the interface requirements and to wait for Flush() to complete.
	c := make(chan error)
	go func() {
		store.wg.Wait()
		syscall.Sync()
		close(c)
	}()
	return c
}

func (store *dirStore) List(prefix string) []string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	files := []string{}
	if prefix == "" {
		prefix = "/"
	}
	// FIXME: should get rid of this hack and use filepath utils instead
	glob := mkpath(store.path, prefix+"x")
	glob = glob[:len(glob)-1]
	if matches, err := filepath.Glob(glob + "*"); err == nil {
		for _, file := range matches {
			if stat, err := os.Stat(file); err == nil && stat.Mode().IsRegular() {
				if s, err := url.QueryUnescape(file); err == nil {
					files = append(files, strings.TrimPrefix(s, store.path+"/"))
				}
			}
		}
	}
	return files
}

type lruItem struct {
	K string
	V Item
}

type lru struct {
	l       *list.List
	m       map[string]*list.Element
	mutex   sync.Mutex
	size    int
	backend Store
}

// Returns LRU cache which is backed up to some other store.
func NewLRU(size int, backend Store) Store {
	return &lru{
		l:       list.New(),
		m:       make(map[string]*list.Element),
		size:    size,
		backend: backend,
	}
}

func (store *lru) Get(key string, item Item) Item {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if el, ok := store.m[key]; ok {
		store.l.MoveToFront(el)
		return el.Value.(*lruItem).V
	} else if store.backend != nil {
		if item := store.backend.Get(key, item); item != nil {
			<-store.put(key, item)
			return item
		}
	}
	return nil
}

func (store *lru) put(key string, item Item) (c <-chan error) {
	if len(store.m) < store.size {
		store.m[key] = store.l.PushFront(&lruItem{key, item})
	} else {
		el := store.l.Back()
		value := el.Value.(*lruItem)
		if store.backend != nil {
			c = store.backend.Set(value.K, value.V)
		}
		delete(store.m, value.K)
		el.Value = &lruItem{key, item}
		store.l.MoveToFront(el)
		store.m[key] = el
	}
	return c
}

func (store *lru) Set(key string, item Item) <-chan error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	c := make(chan error)
	close(c)

	if el, ok := store.m[key]; ok {
		el.Value = &lruItem{key, item}
		store.l.MoveToFront(el)
	} else if item != nil {
		if c := store.put(key, item); c != nil {
			return c
		}
	}
	return c
}

func (store *lru) List(prefix string) []string {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	keys := []string{}
	for k, _ := range store.m {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (store *lru) Flush() <-chan error {
	c := make(chan error)
	if store.backend != nil {
		go func() {
			store.mutex.Lock()
			defer store.mutex.Unlock()
			for _, v := range store.m {
				pair := v.Value.(*lruItem)
				store.backend.Set(pair.K, pair.V)
			}
			if err, ok := <-store.backend.Flush(); ok {
				c <- err
			}
			close(c)
		}()
	}
	return c
}
