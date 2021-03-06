//////////////////////////////////////////////////////////////////////
//
// Given is some code to cache key-value pairs from a database into
// the main memory (to reduce access time). Note that golang's map are
// not entirely thread safe. Multiple readers are fine, but multiple
// writers are not. Change the code to make this thread safe.
//

package main

import (
	"container/list"
)

// CacheSize determines how big the cache can grow
const CacheSize = 100

// KeyStoreCacheLoader is an interface for the KeyStoreCache
type KeyStoreCacheLoader interface {
	// Load implements a function where the cache should gets it's content from
	Load(string) string
}

// KeyStoreCache is a LRU cache for string key-value pairs
type KeyStoreCache struct {
	cache   map[string]string
	pages   list.List
	load    func(string) string
	deletes chan string
}

// New creates a new KeyStoreCache
func New(load KeyStoreCacheLoader) *KeyStoreCache {
	return &KeyStoreCache{
		load:    load.Load,
		cache:   make(map[string]string),
		deletes: make(chan string),
	}
}

// Get gets the key from cache, loads it from the source if needed
func (k *KeyStoreCache) Get(key string) string {
	val, ok := k.cache[key]

	// Miss - load from database and save it in cache
	if !ok {
		k.deletes <- key
	}

	return val
}

func (k *KeyStoreCache) Miss(key string) string {
	val := k.load(key)

	k.pages.PushFront(key)

	// if cache is full remove the least used item
	if len(k.cache) > CacheSize {
		delete(k.cache, k.pages.Back().Value.(string))
		k.pages.Remove(k.pages.Back())
	}

	return val
}

// Loader implements KeyStoreLoader
type Loader struct {
	DB *MockDB
}

// Load gets the data from the database
func (l *Loader) Load(key string) string {
	val, err := l.DB.Get(key)
	if err != nil {
		panic(err)
	}

	return val
}

func main() {
	loader := Loader{
		DB: GetMockDB(),
	}
	cache := New(&loader)

	// Use one goroutine which is the only one which processes "misses"
	// Don't have to use locks that way since no one else will do
	// writes to k.pages or k.cache
	go func() {
		for {
			key := <-cache.deletes
			cache.Miss(key)
		}
	}()

	RunMockServer(cache)
}
