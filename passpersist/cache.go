package passpersist

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
)

func NewCache() *Cache {
	return &Cache{
		staged:    make(map[string]*VarBind),
		committed: make(map[string]*VarBind),
	}
}

type Cache struct {
	staged    map[string]*VarBind
	committed map[string]*VarBind
	index     []string
	mu        sync.RWMutex
}

func (c *Cache) Commit() error {
	c.Lock()
	defer c.Unlock()

	c.committed = c.staged
	c.staged = make(map[string]*VarBind)

	idx := make([]string, 0, len(c.committed))
	for k := range c.committed {
		idx = append(idx, k)
	}
	sort.Strings(idx)
	c.index = idx

	return nil
}

func (c *Cache) Dump() {
	c.RLock()
	defer c.RUnlock()

	out := make(map[string]interface{})
	out["staged"] = c.staged
	out["commited"] = c.committed
	out["index"] = c.index

	y, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(y))
}

func (c *Cache) Get(oid OID) *VarBind {
	c.RLock()
	defer c.RUnlock()

	log.Debug().Msgf("getting value at: %s", oid.String())
	if v, ok := c.committed[oid.String()]; ok {
		return v
	}
	return nil
}

func (c *Cache) GetNext(oid OID) *VarBind {
	c.RLock()
	defer c.RUnlock()

	so := oid.String()
	fo := c.index[0]
	fos, _ := OIDFromString(fo)

	log.Debug().Msgf("getting next value after: %s", so)

	if len(oid) < len(fos) {
		return c.committed[fo]
	}

	for i, o := range c.index {
		if o == so {
			if i < len(c.index) {
				n := c.index[i+1]
				return c.committed[n]
			}
		}
	}
	return nil
}

func (c *Cache) Set(v *VarBind) error {

	log.Debug().Msgf("staging: %s %s %v", v.Oid, v.ValueType, v.Value)

	// // WHY is the oid being overwritten????
	// for k, val := range c.staged {
	// 	fmt.Printf("!!! %s :: %v\n", k, val)
	// }

	c.staged[v.Oid.String()] = v

	// fmt.Printf("??? %+v, %v\n", c.staged[v.Oid.String()], v)

	// // WHY is the oid being overwritten????
	// for k, val := range c.staged {
	// 	fmt.Printf(">>> %s :: %+v\n", k, val)
	// }

	return nil
}

func (c *Cache) Lock() {
	c.mu.Lock()
}

func (c *Cache) RLock() {
	c.mu.RLock()
}

func (c *Cache) Unlock() {
	c.mu.Unlock()
}

func (c *Cache) RUnlock() {
	c.mu.RUnlock()
}
