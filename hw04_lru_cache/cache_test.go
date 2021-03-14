package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	// tests cache.Clear() method logic
	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		c.Clear()

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Equal(t, nil, val)

		val, ok = c.Get("bbb")
		require.False(t, ok)
		require.Equal(t, nil, val)

		wasInCache = c.Set("ccc", 100)
		require.False(t, wasInCache)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 100, val)
	})

	t.Run("respect cache capacity", func(t *testing.T) {
		c := NewCache(3)
		c.Set("aaa", 100)
		c.Set("bbb", 101)
		c.Set("ccc", 102)
		c.Set("ddd", 103)
		el, ok := c.Get("aaa")
		require.False(t, ok)
		require.Equal(t, nil, el)
	})

	t.Run("least recently used item removed", func(t *testing.T) {
		c := NewCache(3)

		// initializing rand with a seed
		rand.Seed(time.Now().UnixNano())
		availableKeys := []Key{"aaa", "bbb", "ccc"}
		// simulate intensive random amount of work with a cache
		randIterationsNum := rand.Intn(10000)
		for i := 0; i < randIterationsNum; i++ {
			c.Set(availableKeys[rand.Intn(len(availableKeys))], rand.Intn(100))
			c.Get(availableKeys[rand.Intn(len(availableKeys))])
		}

		c.Set("ddd", 1234)
		c.Set("eee", 12345)
		c.Set("fff", 123456)

		val, ok := c.Get("aaa")
		require.False(t, ok)
		require.Equal(t, nil, val)

		val, ok = c.Get("bbb")
		require.False(t, ok)
		require.Equal(t, nil, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Equal(t, nil, val)

		val, ok = c.Get("ddd")
		require.True(t, ok)
		require.Equal(t, 1234, val)

		val, ok = c.Get("eee")
		require.True(t, ok)
		require.Equal(t, 12345, val)

		val, ok = c.Get("fff")
		require.True(t, ok)
		require.Equal(t, 123456, val)

	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
