package cache

import (
	"container/list"
	"strconv"
	"sync"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	namespace = "promviz"
	subsystem = "cache"
)

type Cache interface {
	Get(string) *model.Snapshot
	Put(string, *model.Snapshot) bool
	Reset()
}

type Options struct {
	Size int
}

type cacheMetrics struct {
	get  *prometheus.CounterVec
	put  *prometheus.CounterVec
	size prometheus.Gauge
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
		size: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "size",
			Help:      "Current size of cache.",
		}),
	}
	if r != nil {
		r.MustRegister(
			m.get,
			m.put,
			m.size,
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
	items      map[string]*list.Element
}

type item struct {
	key   string
	value *model.Snapshot
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

func (c *cache) Get(key string) (snapshot *model.Snapshot) {
	defer func() {
		if snapshot == nil {
			c.metrics.get.WithLabelValues("miss").Inc()
		} else {
			c.metrics.get.WithLabelValues("hit").Inc()
		}
	}()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[key]; ok {
		c.linkedList.MoveToFront(e)
		snapshot = e.Value.(item).value
		return
	}
	return
}

func (c *cache) Put(key string, snapshot *model.Snapshot) (ok bool) {
	if snapshot == nil {
		return false
	}
	size := 0

	defer func() {
		c.metrics.put.WithLabelValues(strconv.FormatBool(ok)).Inc()
		c.metrics.size.Set(float64(size))
	}()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if e, ok := c.items[key]; ok {
		c.linkedList.MoveToFront(e)
		return true
	}

	element := c.linkedList.PushFront(item{
		key:   key,
		value: snapshot,
	})
	c.items[key] = element

	evict := c.linkedList.Len() > c.options.Size
	if evict {
		e := c.linkedList.Back()
		if e != nil {
			c.linkedList.Remove(e)
			k := e.Value.(item).key
			delete(c.items, k)
		}
	}
	size = len(c.items)
	return true
}

func (c *cache) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.items = make(map[string]*list.Element)
	c.linkedList = list.New()
	c.metrics.size.Set(0)
}
