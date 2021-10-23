// Code generated by github.com/ddouglas/dataloaden, DO NOT EDIT.

package generated

import (
	"context"
	"sync"
	"time"

	"github.com/ddouglas/ledger"
)

// CategoryLoaderConfig captures the config to create a new CategoryLoader
type CategoryLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(ctx context.Context, keys []string) ([]*ledger.PlaidCategory, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewCategoryLoader creates a new CategoryLoader given a fetch, wait, and maxBatch
func NewCategoryLoader(config CategoryLoaderConfig) *CategoryLoader {
	return &CategoryLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// CategoryLoader batches and caches requests
type CategoryLoader struct {
	// this method provides the data for the loader
	fetch func(ctx context.Context, keys []string) ([]*ledger.PlaidCategory, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[string]*ledger.PlaidCategory

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *categoryLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type categoryLoaderBatch struct {
	keys    []string
	data    []*ledger.PlaidCategory
	error   []error
	closing bool
	done    chan struct{}
}

// Load a PlaidCategory by key, batching and caching will be applied automatically
func (l *CategoryLoader) Load(ctx context.Context, key string) (*ledger.PlaidCategory, error) {
	return l.LoadThunk(ctx, key)()
}

// LoadThunk returns a function that when called will block waiting for a PlaidCategory.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *CategoryLoader) LoadThunk(ctx context.Context, key string) func() (*ledger.PlaidCategory, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (*ledger.PlaidCategory, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &categoryLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(ctx, l, key)
	l.mu.Unlock()

	return func() (*ledger.PlaidCategory, error) {
		<-batch.done

		var data *ledger.PlaidCategory
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *CategoryLoader) LoadAll(ctx context.Context, keys []string) ([]*ledger.PlaidCategory, []error) {
	results := make([]func() (*ledger.PlaidCategory, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(ctx, key)
	}

	plaidCategorys := make([]*ledger.PlaidCategory, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		plaidCategorys[i], errors[i] = thunk()
	}
	return plaidCategorys, errors
}

// LoadAllThunk returns a function that when called will block waiting for a PlaidCategorys.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *CategoryLoader) LoadAllThunk(ctx context.Context, keys []string) func() ([]*ledger.PlaidCategory, []error) {
	results := make([]func() (*ledger.PlaidCategory, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(ctx, key)
	}
	return func() ([]*ledger.PlaidCategory, []error) {
		plaidCategorys := make([]*ledger.PlaidCategory, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			plaidCategorys[i], errors[i] = thunk()
		}
		return plaidCategorys, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *CategoryLoader) Prime(key string, value *ledger.PlaidCategory) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
		// and end up with the whole cache pointing to the same value.
		cpy := *value
		l.unsafeSet(key, &cpy)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *CategoryLoader) Clear(key string) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *CategoryLoader) unsafeSet(key string, value *ledger.PlaidCategory) {
	if l.cache == nil {
		l.cache = map[string]*ledger.PlaidCategory{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *categoryLoaderBatch) keyIndex(ctx context.Context, l *CategoryLoader, key string) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(ctx, l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(ctx, l)
		}
	}

	return pos
}

func (b *categoryLoaderBatch) startTimer(ctx context.Context, l *CategoryLoader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(ctx, l)
}

func (b *categoryLoaderBatch) end(ctx context.Context, l *CategoryLoader) {
	b.data, b.error = l.fetch(ctx, b.keys)
	close(b.done)
}
