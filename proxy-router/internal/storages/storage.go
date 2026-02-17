package storages

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	badger "github.com/dgraph-io/badger/v4"
)

// Default configuration values for BadgerDB
const (
	DefaultGCInterval      = 5 * time.Minute
	DefaultGCRatio         = 0.5
	DefaultMetricsInterval = 5 * time.Minute
)

type Storage struct {
	db     *badger.DB
	log    lib.ILogger
	stopGC chan struct{}
	gcDone chan struct{}
}

// NewStorage opens a BadgerDB at the given path and starts background GC.
// Returns an error instead of calling log.Fatal so callers can handle failures gracefully.
func NewStorage(log lib.ILogger, path string) (*Storage, error) {
	storageLogger := NewStorageLogger(log)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create storage directory %s: %w", path, err)
	}

	opts := badger.DefaultOptions(path)
	opts.Logger = storageLogger
	opts.NumVersionsToKeep = 1
	opts.CompactL0OnClose = true

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db at %s: %w", path, err)
	}

	s := &Storage{
		db:     db,
		log:    log.Named("BADGER"),
		stopGC: make(chan struct{}),
		gcDone: make(chan struct{}),
	}
	s.startGC(db, DefaultGCRatio, DefaultGCInterval)
	s.startMetrics(DefaultMetricsInterval)

	return s, nil
}

// NewTestStorage creates an in-memory storage for testing.
func NewTestStorage() *Storage {
	opts := badger.DefaultOptions("")
	opts.InMemory = true
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Storage{db: db, stopGC: make(chan struct{}), gcDone: make(chan struct{})}
}

// Close stops GC, flushes data, and closes the underlying BadgerDB.
func (s *Storage) Close() error {
	if s.stopGC != nil {
		close(s.stopGC)
		if s.gcDone != nil {
			<-s.gcDone
		}
	}
	if err := s.db.Close(); err != nil {
		if s.log != nil {
			s.log.Errorf("error closing badger db: %s", err)
		}
		return fmt.Errorf("error closing badger db: %w", err)
	}
	return nil
}

// startGC runs periodic value log garbage collection in the background.
func (s *Storage) startGC(db *badger.DB, discardRatio float64, interval time.Duration) {
	go func() {
		defer close(s.gcDone)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopGC:
				return
			case <-ticker.C:
				gcRuns := 0
				for {
					err := db.RunValueLogGC(discardRatio)
					if err != nil {
						if !errors.Is(err, badger.ErrNoRewrite) {
							s.log.Warnf("badger GC error: %s", err)
						}
						break
					}
					gcRuns++
				}

				lsmSize, vlogSize := db.Size()
				totalMB := float64(lsmSize+vlogSize) / (1024 * 1024)
				if gcRuns > 0 {
					s.log.Infof("badger GC completed: %d cycles, lsm=%.1fMB, vlog=%.1fMB, total=%.1fMB",
						gcRuns, float64(lsmSize)/(1024*1024), float64(vlogSize)/(1024*1024), totalMB)
				} else {
					s.log.Infof("badger GC: ok, nothing to clean, lsm=%.1fMB, vlog=%.1fMB, total=%.1fMB",
						float64(lsmSize)/(1024*1024), float64(vlogSize)/(1024*1024), totalMB)
				}
			}
		}
	}()
}

// startMetrics periodically logs database size metrics.
func (s *Storage) startMetrics(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopGC:
				return
			case <-ticker.C:
				lsmSize, vlogSize := s.db.Size()
				totalMB := float64(lsmSize+vlogSize) / (1024 * 1024)
				if totalMB > 10000 {
					s.log.Warnf("badger DB size is large: lsm=%.1fMB, vlog=%.1fMB, total=%.1fMB",
						float64(lsmSize)/(1024*1024), float64(vlogSize)/(1024*1024), totalMB)
				}
			}
		}
	}()
}

// HealthCheck verifies the database is operational by doing a test write/read/delete cycle.
func (s *Storage) HealthCheck() error {
	testKey := []byte("_health_check")
	testVal := []byte(time.Now().Format(time.RFC3339Nano))

	if err := s.Set(testKey, testVal); err != nil {
		return fmt.Errorf("badger health check write failed: %w", err)
	}

	readVal, err := s.Get(testKey)
	if err != nil {
		return fmt.Errorf("badger health check read failed: %w", err)
	}
	if string(readVal) != string(testVal) {
		return fmt.Errorf("badger health check mismatch: wrote %q, read %q", testVal, readVal)
	}

	if err := s.Delete(testKey); err != nil {
		return fmt.Errorf("badger health check delete failed: %w", err)
	}

	return nil
}

// DBSize returns the LSM and value log sizes in bytes.
func (s *Storage) DBSize() (lsmSize int64, vlogSize int64) {
	return s.db.Size()
}

// Get retrieves the value for the given key. Returns badger.ErrKeyNotFound if the key does not exist.
func (s *Storage) Get(key []byte) ([]byte, error) {
	var valCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return valCopy, err
}

// GetPrefix returns all keys matching the given prefix.
func (s *Storage) GetPrefix(prefix []byte) ([][]byte, error) {
	keys := make([][]byte, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: false, Prefix: prefix})
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			keys = append(keys, k)
		}
		return nil
	})
	return keys, err
}

// GetPrefixWithValues returns all key-value pairs matching the given prefix in a single transaction.
// Supports context cancellation for long-running scans.
func (s *Storage) GetPrefixWithValues(prefix []byte, ctxOpts ...context.Context) ([][]byte, [][]byte, error) {
	var ctx context.Context
	if len(ctxOpts) > 0 && ctxOpts[0] != nil {
		ctx = ctxOpts[0]
	}

	var keys [][]byte
	var values [][]byte
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: true, PrefetchSize: 100, Prefix: prefix})
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if ctx != nil {
				if err := ctx.Err(); err != nil {
					return fmt.Errorf("scan cancelled: %w", err)
				}
			}
			item := it.Item()
			k := item.KeyCopy(nil)
			v, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("error reading value for key %s: %w", string(k), err)
			}
			keys = append(keys, k)
			values = append(values, v)
		}
		return nil
	})
	return keys, values, err
}

// Set stores a key-value pair.
func (s *Storage) Set(key, val []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

// SetWithTTL stores a key-value pair that expires after the given duration.
func (s *Storage) SetWithTTL(key, val []byte, ttl time.Duration) error {
	return s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, val).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// Delete removes a key from the database.
func (s *Storage) Delete(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// RunInTransaction executes a function within a single read-write BadgerDB transaction.
// This is used for atomic read-modify-write operations.
func (s *Storage) RunInTransaction(fn func(txn *badger.Txn) error) error {
	return s.db.Update(fn)
}

// Paginate returns keys matching the prefix with cursor-based pagination.
func (s *Storage) Paginate(prefix []byte, cursor []byte, limit uint) ([][]byte, []byte, error) {
	keys := make([][]byte, 0, limit)
	var nextCursor []byte

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: false, Prefix: prefix})
		defer it.Close()

		if len(cursor) == 0 {
			cursor = prefix
		}

		for it.Seek(cursor); it.ValidForPrefix(prefix); it.Next() {
			if uint(len(keys)) >= limit {
				nextCursor = it.Item().KeyCopy(nil)
				break
			}
			item := it.Item()
			k := item.KeyCopy(nil)
			keys = append(keys, k)
		}
		return nil
	})
	return keys, nextCursor, err
}
