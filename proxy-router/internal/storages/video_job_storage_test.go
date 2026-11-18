package storages

import (
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func TestVideoTaskRoundTrip(t *testing.T) {
	ss := NewSessionStorage(NewTestStorage())

	record := []byte(`{"vendor":"byteplus","task_id":"cgt-1"}`)
	require.NoError(t, ss.PutVideoTask("0xABCDEF", "vjob_1", record))

	got, err := ss.GetVideoTask("0xABCDEF", "vjob_1")
	require.NoError(t, err)
	require.Equal(t, record, got)

	// The session id is a hex hash: the key is case-insensitive, so a
	// checksummed and a lowercased spelling read the same record.
	got, err = ss.GetVideoTask("0xabcdef", "vjob_1")
	require.NoError(t, err)
	require.Equal(t, record, got)

	// Put overwrites the same job's record.
	updated := []byte(`{"vendor":"byteplus","task_id":"cgt-2"}`)
	require.NoError(t, ss.PutVideoTask("0xabcdef", "vjob_1", updated))
	got, err = ss.GetVideoTask("0xABCDEF", "vjob_1")
	require.NoError(t, err)
	require.Equal(t, updated, got)
}

// A session can carry several jobs: GetVideoTasks lists that session's
// records and no other session's.
func TestGetVideoTasksListsOneSession(t *testing.T) {
	storage := NewTestStorage()
	ss := NewSessionStorage(storage)

	require.NoError(t, ss.PutVideoTask("0xAB", "vjob_1", []byte(`"one"`)))
	require.NoError(t, ss.PutVideoTask("0xab", "vjob_2", []byte(`"two"`)))
	require.NoError(t, ss.PutVideoTask("0xabc", "vjob_3", []byte(`"other session"`)))
	require.NoError(t, storage.Set([]byte("video_task:0xab"), []byte(`"bare"`)))

	tasks, err := ss.GetVideoTasks("0xAB")
	require.NoError(t, err)
	got := make([]string, 0, len(tasks))
	for _, task := range tasks {
		got = append(got, string(task))
	}
	require.ElementsMatch(t, []string{`"one"`, `"two"`}, got)

	none, err := ss.GetVideoTasks("0xdead")
	require.NoError(t, err)
	require.Empty(t, none)
}

func TestVideoJobRoundTrip(t *testing.T) {
	ss := NewSessionStorage(NewTestStorage())

	record := []byte(`{"id":"vj_1","status":"queued"}`)
	require.NoError(t, ss.PutVideoJob("vj_1", record))

	got, err := ss.GetVideoJobs()
	require.NoError(t, err)
	require.Equal(t, [][]byte{record}, got)

	// Put overwrites: one record per job.
	updated := []byte(`{"id":"vj_1","status":"succeeded"}`)
	require.NoError(t, ss.PutVideoJob("vj_1", updated))
	got, err = ss.GetVideoJobs()
	require.NoError(t, err)
	require.Equal(t, [][]byte{updated}, got)
}

func TestVideoRecordsMissingKeyIsNilNil(t *testing.T) {
	ss := NewSessionStorage(NewTestStorage())

	task, err := ss.GetVideoTask("0xdead", "vjob_1")
	require.NoError(t, err)
	require.Nil(t, task)

	jobs, err := ss.GetVideoJobs()
	require.NoError(t, err)
	require.Empty(t, jobs)
}

// The task and job records share a namespace with sessions and users; the
// job listing must return job records only.
func TestGetVideoJobsListsOnlyVideoJobKeys(t *testing.T) {
	storage := NewTestStorage()
	ss := NewSessionStorage(storage)

	require.NoError(t, ss.PutVideoJob("vj_a", []byte(`"a"`)))
	require.NoError(t, ss.PutVideoJob("vj_b", []byte(`"b"`)))
	require.NoError(t, ss.PutVideoTask("0x01", "vjob_1", []byte(`"task"`)))

	// Neighbours whose keys share the "video_job" spelling but not the
	// "video_job:" prefix.
	require.NoError(t, storage.Set([]byte("video_jobs:x"), []byte(`"plural"`)))
	require.NoError(t, storage.Set([]byte("video_job"), []byte(`"bare"`)))
	require.NoError(t, storage.Set([]byte("session:0x01"), []byte(`"session"`)))

	jobs, err := ss.GetVideoJobs()
	require.NoError(t, err)

	got := make([]string, 0, len(jobs))
	for _, j := range jobs {
		got = append(got, string(j))
	}
	require.ElementsMatch(t, []string{`"a"`, `"b"`}, got)
}

// Both records expire after VideoJobTTL so finished jobs do not pile up.
func TestVideoRecordsExpireAfterVideoJobTTL(t *testing.T) {
	storage := NewTestStorage()
	ss := NewSessionStorage(storage)

	require.NoError(t, ss.PutVideoTask("0xAB", "vjob_1", []byte(`"task"`)))
	require.NoError(t, ss.PutVideoJob("vj_ttl", []byte(`"job"`)))

	want := time.Now().Add(VideoJobTTL)
	for _, key := range []string{"video_task:0xab:vjob_1", "video_job:vj_ttl"} {
		var expiresAt uint64
		require.NoError(t, storage.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			expiresAt = item.ExpiresAt()
			return nil
		}), key)
		require.NotZero(t, expiresAt, "%s has no TTL", key)
		require.WithinDuration(t, want, time.Unix(int64(expiresAt), 0), time.Minute, key)
	}
}
