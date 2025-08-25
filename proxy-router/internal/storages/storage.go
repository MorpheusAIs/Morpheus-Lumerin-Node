package storages

import (
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	badger "github.com/dgraph-io/badger/v4"
)

type Storage struct {
	db *badger.DB
	stopGC chan struct{}
	gcDone chan struct{}
}

func NewStorage(log lib.ILogger, path string) *Storage {
	storageLogger := NewStorageLogger(log)
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		storageLogger.Debugf("%s", err)
	}
	opts := badger.DefaultOptions(path)
	opts.Logger = storageLogger

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	s := &Storage{db: db, stopGC: make(chan struct{}), gcDone: make(chan struct{})}
	go func() {
		defer close(s.gcDone)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopGC:
				return
			case <-ticker.C:
				for {
					if err := db.RunValueLogGC(0.7); err != nil {
						break
					}
				}
			}
		}
	}()
	return s
}

func NewTestStorage() *Storage {
	opts := badger.DefaultOptions("")
	opts.InMemory = true
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Storage{db: db, stopGC: make(chan struct{}), gcDone: make(chan struct{})}
}

func (s *Storage) Close() {
	if s.stopGC != nil {
		close(s.stopGC)
		if s.gcDone != nil {
			<-s.gcDone
		}
	}
	s.db.Close()
}

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

func (s *Storage) Set(key, val []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (s *Storage) Delete(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

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
