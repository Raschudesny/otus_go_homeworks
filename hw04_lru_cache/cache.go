package hw04lrucache

import (
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool // Add value in cache using key
	Get(key Key) (interface{}, bool)     // Get value from cache using key
	Clear()                              // Clear cache
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*listItem
	rwMutex  sync.RWMutex
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()

	if val, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(val)
		cache.queue.Front().Value = cacheItem{key, value}
		cache.items[key] = cache.queue.Front()
		return ok
	}

	if cache.capacity == cache.queue.Len() {
		leastRecentlyUsedItem := cache.queue.Back()
		displacedCachedItem, isCacheItem := leastRecentlyUsedItem.Value.(cacheItem)
		if isCacheItem {
			cache.queue.Remove(leastRecentlyUsedItem)
			delete(cache.items, displacedCachedItem.key)
		} else {
			// TODO Right now just fairly panic here, because can't break interface contract
			// TODO func (cache *lruCache) Set(key Key, value interface{}) bool
			// TODO maybe Set() method should return (bool, error) ?
			panic("Internal lruCache implementation error occurred")
		}
	}
	item := cacheItem{key, value}
	cache.items[key] = cache.queue.PushFront(item)
	return false
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()

	if val, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(val)
		val, ok := val.Value.(cacheItem)
		if ok {
			return val.value, true
		}
		// TODO Right now just fairly panic here, because can't break interface contract
		// TODO func (cache *lruCache) Get(key Key) (interface{}, bool)
		// TODO maybe Get() method should return (interface{}, bool, error) ?
		panic("Internal lruCache implementation error occurred")
	} else {
		return nil, false
	}
}

func (cache *lruCache) Clear() {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.items = make(map[Key]*listItem, cache.capacity)
	for cache.queue.Front() != nil {
		cache.queue.Remove(cache.queue.Front())
	}
}

type cacheItem struct {
	// Place your code here
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{capacity: capacity, queue: NewList(), items: make(map[Key]*listItem, capacity), rwMutex: sync.RWMutex{}}
}
