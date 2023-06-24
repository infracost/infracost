package apiclient

import (
	"sync"

	hashstructure "github.com/mitchellh/hashstructure/v2"
	"github.com/tidwall/gjson"
)

type Cache struct {
	cache map[uint64]gjson.Result
	mtx   sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[uint64]gjson.Result),
		mtx:   sync.RWMutex{},
	}
}

func (c *Cache) Add(query GraphQLQuery, result gjson.Result) {
	hash, _ := hashstructure.Hash(query, hashstructure.FormatV2, nil)

	c.mtx.Lock()
	c.cache[hash] = result
	c.mtx.Unlock()
}

func (c *Cache) Lookup(query GraphQLQuery) (gjson.Result, bool) {
	hash, _ := hashstructure.Hash(query, hashstructure.FormatV2, nil)

	c.mtx.RLock()
	r, ok := c.cache[hash]
	c.mtx.RUnlock()
	return r, ok
}
