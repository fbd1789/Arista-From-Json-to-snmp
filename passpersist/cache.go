package passpersist

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
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
	index     OIDs
}

func (c *Cache) getIndex(o OID) (int, error) {
	for p, v := range c.index {
		if v.Equal(o) {
			return p, nil
		}
	}

	for p, v := range c.index {
		if v.StartsWith(o) {
			return p - 1, nil
		}
	}
	return 0, errors.New("OID or prefix does not exist")
}

func (c *Cache) Commit() error {
	c.Lock()
	defer c.Unlock()

	c.committed = c.staged
	c.staged = make(map[string]*VarBind)

	idx := make(OIDs, 0, len(c.committed))
	for _, vb := range c.committed {
		idx = append(idx, vb.OID)
	}

	idx = idx.Sort()
	c.index = idx

	return nil
}

func (c *Cache) DumpIndex() {
	c.RLock()
	defer c.RUnlock()

	y, _ := json.MarshalIndent(c.index, "", "  ")
	fmt.Println(string(y))
}

func (c *Cache) Dump() {
	c.RLock()
	defer c.RUnlock()

	o, _ := json.MarshalIndent(c.committed, "", "  ")
	fmt.Println(string(o))
}

func (c *Cache) Get(oid OID) *VarBind {
	c.RLock()
	defer c.RUnlock()

	slog.Debug("getting value at oid", "oid", oid.String())
	if v, ok := c.committed[oid.String()]; ok {
		slog.Debug("got value", "oid", oid.String(), "value", v.Value.String())
		return v
	}
	return nil
}

func (c *Cache) GetNext(oid OID) *VarBind {
	c.RLock()
	defer c.RUnlock()

	slog.Debug("getting next value after", "oid", oid.String())

	idx, err := c.getIndex(oid)
	slog.Debug("got index at", "oid", oid.String(), "index", idx)
	if err != nil {
		slog.Warn("failed to get index", slog.Any("error", err.Error()), "oid", oid.String())
		return nil
	}

	idx++

	if idx < len(c.index) {
		next := c.index[idx]
		if v, ok := c.committed[next.String()]; ok {
			return v
		} else {
			//
		}
	}

	return nil
}

func (c *Cache) Set(v *VarBind) error {
	c.Lock()
	defer c.Unlock()

	slog.Debug("staging", "oid", v.OID, "type", v.ValueType, "value", v.Value)

	c.staged[v.OID.String()] = v

	return nil
}
