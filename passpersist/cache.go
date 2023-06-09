package passpersist

import (
	"encoding/json"
	"fmt"
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
	sync.RWMutex
	staged    map[string]*VarBind
	committed map[string]*VarBind
	index     Oids
}

func (c *Cache) getIndex(o *Oid) int {
	for p, v := range c.index {
		if v.Equal(o) {
			return p
		}
	}
	return -1
}

func (c *Cache) Commit() error {
	c.Lock()
	defer c.Unlock()

	c.committed = c.staged
	c.staged = make(map[string]*VarBind)

	idx := make(Oids, 0, len(c.committed))
	for _, vb := range c.committed {
		idx = append(idx, vb.Oid)
	}
	//sort.Strings(idx)
	idx = idx.Sort()
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

func (c *Cache) Get(oid *Oid) *VarBind {
	c.RLock()
	defer c.RUnlock()

	log.Debug().Msgf("getting value at: %s", oid.String())
	if v, ok := c.committed[oid.String()]; ok {
		log.Debug().Msgf("got value at: %s=%s", oid.String(), &v.Value)
		return v
	}
	return nil
}

func (c *Cache) GetNext(oid *Oid) *VarBind {
	c.RLock()
	defer c.RUnlock()

	log.Debug().Msgf("getting next value after: %s", oid.String())

	idx := c.getIndex(oid)
	if idx < 0 {
		if !c.index[0].Contains(oid) {
			return nil
		}
	}

	idx++

	if idx < len(c.index) {
		next := c.index[idx]
		if v, ok := c.committed[next.String()]; ok {
			return v
		}
	}

	return nil
}

func (c *Cache) Set(v *VarBind) error {
	c.Lock()
	defer c.Unlock()

	log.Debug().Msgf("staging: %s %s %v", v.Oid, v.ValueType, v.Value)

	c.staged[v.Oid.String()] = v

	return nil
}
