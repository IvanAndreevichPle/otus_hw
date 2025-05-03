package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

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

	t.Run("purge logic", func(t *testing.T) {
		t.Run("purge on capacity overflow", func(t *testing.T) {
			c := NewCache(3)
			c.Set("a", 1)
			c.Set("b", 2)
			c.Set("c", 3)
			c.Set("d", 4)

			_, ok := c.Get("a")
			require.False(t, ok)
		})

		t.Run("purge least recently used", func(t *testing.T) {
			c := NewCache(3)
			c.Set("a", 1)
			c.Set("b", 2)
			c.Set("c", 3)

			c.Get("a")
			c.Get("c")
			c.Set("d", 4)

			_, ok := c.Get("b")
			require.False(t, ok)
		})

		t.Run("update existing item", func(t *testing.T) {
			c := NewCache(3)
			c.Set("a", 1)
			c.Set("b", 2)
			c.Set("c", 3)
			c.Set("a", 4)
			c.Set("d", 5)

			_, ok := c.Get("b")
			require.False(t, ok)
		})
	})
}

func TestCacheMultithreading(_ *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000))))
		}
	}()

	wg.Wait()
}

func TestCacheMultithreadingMixed(_ *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(100))))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				c.Clear()
			}
		}
	}()

	wg.Wait()
}

func TestCacheEdgeCases(t *testing.T) {
	t.Run("negative capacity", func(t *testing.T) {
		c := NewCache(-1)
		c.Set("a", 1)
		_, ok := c.Get("a")
		require.False(t, ok)
	})

	t.Run("nil value handling", func(t *testing.T) {
		c := NewCache(1)
		c.Set("a", nil)
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Nil(t, val)
	})

	t.Run("empty string key", func(t *testing.T) {
		c := NewCache(1)
		c.Set("", 1)
		val, ok := c.Get("")
		require.True(t, ok)
		require.Equal(t, 1, val)
	})

	t.Run("large values", func(t *testing.T) {
		c := NewCache(2)
		largeSlice := make([]int, 1000000)
		c.Set("large", largeSlice)
		val, ok := c.Get("large")
		require.True(t, ok)
		require.Equal(t, largeSlice, val)
	})
}

func TestCacheStressTest(_ *testing.T) {
	c := NewCache(100)
	wg := &sync.WaitGroup{}
	ops := 1000
	goroutines := 10
	wg.Add(goroutines * 3)

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := Key(strconv.Itoa(rand.Intn(ops)))
				c.Set(key, j)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := Key(strconv.Itoa(rand.Intn(ops)))
				c.Get(key)
			}
		}(i)
	}

	// Mixed operations
	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				switch rand.Intn(3) {
				case 0:
					key := Key(strconv.Itoa(rand.Intn(ops)))
					c.Set(key, j)
				case 1:
					key := Key(strconv.Itoa(rand.Intn(ops)))
					c.Get(key)
				case 2:
					c.Clear()
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestCacheConsistency(t *testing.T) {
	t.Run("value overwrite consistency", func(t *testing.T) {
		c := NewCache(2)
		c.Set("a", 1)
		c.Set("a", 2)
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 2, val)
	})

	t.Run("capacity consistency", func(t *testing.T) {
		c := NewCache(2)
		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)
		_, ok := c.Get("a")
		require.False(t, ok)
		val, ok := c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)
	})

	t.Run("clear consistency", func(t *testing.T) {
		c := NewCache(2)
		c.Set("a", 1)
		c.Clear()
		c.Set("b", 2)
		_, ok := c.Get("a")
		require.False(t, ok)
		val, ok := c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)
	})
}
