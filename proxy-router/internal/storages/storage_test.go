package storages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func TestNewStorage_CorruptionRecovery(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "testdb")
	log := &lib.LoggerMock{}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	require.NoError(t, err)

	for i := 0; i < 5000; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			key := []byte(fmt.Sprintf("key-%06d", i))
			val := make([]byte, 4096)
			return txn.Set(key, val)
		})
		require.NoError(t, err)
	}

	require.NoError(t, db.Close())

	// Place a non-BadgerDB file in the data directory
	foreignFile := filepath.Join(dbPath, "config.json")
	require.NoError(t, os.WriteFile(foreignFile, []byte(`{"keep":"me"}`), 0644))

	sstFiles, err := filepath.Glob(filepath.Join(dbPath, "*.sst"))
	require.NoError(t, err)
	require.NotEmpty(t, sstFiles, "expected SST files after writing data")

	t.Logf("created %d SST files, deleting one to simulate corruption", len(sstFiles))
	require.NoError(t, os.Remove(sstFiles[0]))

	storage, err := NewStorage(log, dbPath)
	require.NoError(t, err, "NewStorage should recover from corruption")
	defer storage.Close()

	require.NoError(t, storage.Set([]byte("post-recovery"), []byte("works")))
	val, err := storage.Get([]byte("post-recovery"))
	require.NoError(t, err)
	require.Equal(t, "works", string(val))

	_, err = storage.Get([]byte("key-000001"))
	require.ErrorIs(t, err, badger.ErrKeyNotFound, "old data should be gone after recovery")

	// Non-BadgerDB file must survive the recovery
	content, err := os.ReadFile(foreignFile)
	require.NoError(t, err, "non-BadgerDB file should not be deleted during recovery")
	require.Equal(t, `{"keep":"me"}`, string(content))
}

func TestNewStorage_NormalOpen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "testdb")
	log := &lib.LoggerMock{}

	storage, err := NewStorage(log, dbPath)
	require.NoError(t, err)

	require.NoError(t, storage.Set([]byte("hello"), []byte("world")))
	require.NoError(t, storage.Close())

	storage2, err := NewStorage(log, dbPath)
	require.NoError(t, err)
	defer storage2.Close()

	val, err := storage2.Get([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, "world", string(val))
}

func TestIsBadgerFile(t *testing.T) {
	badger := []string{"MANIFEST", "KEYREGISTRY", "DISCARD", "LOCK", "000001.sst", "000002.vlog", "00001.mem"}
	for _, name := range badger {
		require.True(t, isBadgerFile(name), "expected %q to be recognized as a BadgerDB file", name)
	}

	nonBadger := []string{"config.json", ".env", "readme.md", "data.db", "notes.txt"}
	for _, name := range nonBadger {
		require.False(t, isBadgerFile(name), "expected %q to NOT be recognized as a BadgerDB file", name)
	}
}

func TestIsBadgerCorruptionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"generic error", fmt.Errorf("connection refused"), false},
		{"permission denied", fmt.Errorf("permission denied"), false},
		{"file does not exist for table", fmt.Errorf("file does not exist for table 8"), true},
		{"file with ID not found", fmt.Errorf("file with ID: 2719 not found"), true},
		{"MANIFEST error", fmt.Errorf("MANIFEST file corrupted"), true},
		{"checksum mismatch", fmt.Errorf("checksum mismatch for table"), true},
		{"table file corruption", fmt.Errorf("Table file corruption detected"), true},
		{"value log truncate", fmt.Errorf("Value log truncate required"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isBadgerCorruptionError(tt.err))
		})
	}
}
