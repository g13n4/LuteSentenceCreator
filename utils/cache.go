package utils

func NewCache[K, V comparable](size int) *Cache[K, V] {
	data := map[K]V{}
	queue := NewCircleQueue[K](size)
	return &Cache[K, V]{Size: size, cache: data, queue: *queue}
}

type Cache[K, V comparable] struct {
	Size  int
	cache map[K]V
	queue CircleQueue[K]
}

func (c *Cache[K, V]) ToCache(key K, val V) {
	var emptyVar K
	_, ok := c.cache[key]
	if ok {
		return
	}
	oldVal := c.queue.Add(key)

	if oldVal != emptyVar {
		delete(c.cache, oldVal)
	}

	c.cache[key] = val
	return
}

func (c *Cache[K, V]) FromCache(key K) (V, bool) {
	v, ok := c.cache[key]
	return v, ok
}
