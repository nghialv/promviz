package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Cache interface {
	Get(time.Time) *model.GraphData
	Put(*model.GraphData) bool
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
	items     map[time.Time]*list.Element
}

func NewCache(logger *zap.Logger, r prometheus.Registerer, opts *Options) Cache {
	c := &lru{
		logger:  logger,
		options: opts,
	}
	c.Reset()
	return c
}

func (c *lru) Get(t time.Time) *model.GraphData {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[t]; ok {
		c.evictList.MoveToFront(e)
		return e.Value.(*model.GraphData)
	}
	return nil
}

func (c *lru) Put(gd *model.GraphData) bool {
	if gd == nil {
		return false
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[gd.Time]; ok {
		c.evictList.MoveToFront(e)
		return false
	}

	element := c.evictList.PushFront(gd)
	c.items[gd.Time] = element

	evict := c.evictList.Len() > c.options.Size
	if evict {
		e := c.evictList.Back()
		if e != nil {
			c.evictList.Remove(e)
			old := e.Value.(*model.GraphData)
			delete(c.items, old.Time)
		}
	}
	return evict
}

func (c *lru) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.items = make(map[time.Time]*list.Element)
	c.evictList = list.New()
}
