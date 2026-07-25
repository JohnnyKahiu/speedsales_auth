package localcache
package localcache

import (
	"context"
	"io"
	"os"
	"sync"
	"time"
)

// Serializer allows users to inject custom encoders (JSON, Gob, MessagePack).
type Serializer[K comparable, V any] interface {
	Marshal(data map[K]V, w io.Writer) error
	Unmarshal(r io.Reader, target *map[K]V) error
}

type Cache[K comparable, V any] struct {
	mu         sync.RWMutex
	data       map[K]V
	filePath   string
	serializer Serializer[K, V]
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// New creates an instance of your generic cache, loads past state, and spins up a background flusher.
func New[K comparable, V any](filePath string, interval time.Duration, s Serializer[K, V]) (*Cache[K, V], error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache[K, V]{
		data:       make(map[K]V),
		filePath:   filePath,
		serializer: s,
		ctx:        ctx,
		cancel:     cancel,
	}

	if err := c.loadFromFile(); err != nil {
		return nil, err
	}

	// Active background ticker
	c.wg.Add(1)
	go c.startBackgroundFlush(interval)

	return c, nil
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.data[key]
	return val, exists
}

func (c *Cache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}

func (c *Cache[K, V]) SaveToFile() error {
	tempPath := c.filePath + ".tmp"
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// Deep copy data under an RLock to keep disk I/O outside critical mutex section
	c.mu.RLock()
	err = c.serializer.Marshal(c.data, file)
	c.mu.RUnlock()
	file.Close()

	if err != nil {
		os.Remove(tempPath)
		return err
	}

	return os.Rename(tempPath, c.filePath)
}

func (c *Cache[K, V]) startBackgroundFlush(interval time.Duration) {
	defer c.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = c.SaveToFile()
		case <-c.ctx.Done():
			_ = c.SaveToFile() // Final flush on exit
			return
		}
	}
}

func (c *Cache[K, V]) loadFromFile() error {
	file, err := os.Open(c.filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serializer.Unmarshal(file, &c.data)
}

// Close ensures a graceful shutdown, preventing file corruption or missed entries.
func (c *Cache[K, V]) Close() {
	c.cancel()
	c.wg.Wait()
}

type JSONSerializer[K comparable, V any] struct{}

func (j JSONSerializer[K, V]) Marshal(data map[K]V, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (j JSONSerializer[K, V]) Unmarshal(r io.Reader, target *map[K]V) error {
	return json.NewDecoder(r).Decode(target)
}
Use code with caution.4. How Someone Consumes Your Reusable PackageOnce published, another developer can use your library type-safely like this:gopackage main

import (
	"fmt"
	"time"
	"yourproject/localcache"
)

type UserProfile struct {
	Name string
	Age  int
}
