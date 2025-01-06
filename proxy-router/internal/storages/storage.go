package storages

import (
	"os"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	badger "github.com/dgraph-io/badger/v4"
)

type Storage struct {
	db *badger.DB
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

	return &Storage{db}
}

func NewTestStorage() *Storage {
	opts := badger.DefaultOptions("")
	opts.InMemory = true
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Storage{db}
}

func (s *Storage) Close() {
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
