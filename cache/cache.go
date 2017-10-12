package cache

import (
	"container/list"
	"strconv"
	"sync"

	"github.com/nghialv/promviz/storage"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	namespace = "promviz"
	subsystem = "cache"
)

type Cache interface {
	Get(int64) *storage.Chunk
	Put(int64, *storage.Chunk) bool
	Reset()
}

type Options struct {
	Size int
}

type cacheMetrics struct {
	get    *prometheus.CounterVec
	put    *prometheus.CounterVec
	length prometheus.Gauge
}

func newCacheMetrics(r prometheus.Registerer) *cacheMetrics {
	m := &cacheMetrics{
		get: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "get",
			Help:      "Total number of Get requests.",
		},
			[]string{"status"},
		),
		put: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "put",
			Help:      "Total number of Put requests.",
		},
			[]string{"status"},
		),
		length: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "length",
			Help:      "Current length of cache.",
		}),
	}
	if r != nil {
		r.MustRegister(
			m.get,
			m.put,
			m.length,
		)
	}
	return m
}

type cache struct {
	logger  *zap.Logger
	metrics *cacheMetrics

	options *Options
	mtx     sync.Mutex

	linkedList *list.List
	items      map[int64]*list.Element
}

type item struct {
	key   int64
	value *storage.Chunk
}

func NewCache(logger *zap.Logger, r prometheus.Registerer, opts *Options) Cache {
	c := &cache{
		logger:  logger,
		metrics: newCacheMetrics(r),
		options: opts,
	}
	c.Reset()
	return c
}

func (c *cache) Get(chunkID int64) (chunk *storage.Chunk) {
	defer func() {
		if chunk == nil {
			c.metrics.get.WithLabelValues("miss").Inc()
		} else {
			c.metrics.get.WithLabelValues("hit").Inc()
		}
	}()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[chunkID]; ok {
		c.linkedList.MoveToFront(e)
		chunk = e.Value.(item).value
		return
	}
	return
}

func (c *cache) Put(chunkID int64, chunk *storage.Chunk) (ok bool) {
	if chunk == nil {
		return false
	}
	length := 0

	defer func() {
		c.metrics.put.WithLabelValues(strconv.FormatBool(ok)).Inc()
		c.metrics.length.Set(float64(length))
	}()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[chunkID]; ok {
		c.linkedList.MoveToFront(e)
		return true
	}

	element := c.linkedList.PushFront(item{
		key:   chunkID,
		value: chunk,
	})
	c.items[chunkID] = element

	evict := c.linkedList.Len() > c.options.Size
	if evict {
		e := c.linkedList.Back()
		if e != nil {
			c.linkedList.Remove(e)
			k := e.Value.(item).key
			delete(c.items, k)
		}
	}
	length = len(c.items)
	return true
}

func (c *cache) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.items = make(map[int64]*list.Element)
	c.linkedList = list.New()
	c.metrics.length.Set(0)
}
