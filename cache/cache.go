package cache

import (
	"container/list"
	"sync"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Cache interface {
	Get(string) *model.Snapshot
	Put(string, *model.Snapshot) bool
	Reset()
}

type Options struct {
	Size int
}

type lru struct {
	logger  *zap.Logger
	options *Options
	mtx     sync.Mutex

	evictList *list.List
	items     map[string]*list.Element
}

func NewCache(logger *zap.Logger, r prometheus.Registerer, opts *Options) Cache {
	c := &lru{
		logger:  logger,
		options: opts,
	}
	c.Reset()
	return c
}

func (c *lru) Get(key string) *model.Snapshot {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[key]; ok {
		c.evictList.MoveToFront(e)
		return e.Value.(*model.Snapshot)
	}
	return nil
}

func (c *lru) Put(key string, snapshot *model.Snapshot) bool {
	if snapshot == nil {
		return false
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[key]; ok {
		c.evictList.MoveToFront(e)
		return false
	}

	element := c.evictList.PushFront(snapshot)
	c.items[key] = element

	evict := c.evictList.Len() > c.options.Size
	// if evict {
	// 	e := c.evictList.Back()
	// 	if e != nil {
	// 		c.evictList.Remove(e)
	// 		old := e.Value.(*model.Snapshot)
	// 		delete(c.items, old.Time)
	// 	}
	// }
	return evict
}

func (c *lru) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList = list.New()
}
