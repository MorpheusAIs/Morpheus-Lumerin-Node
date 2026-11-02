package storages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// VideoJobTTL keeps video job records a little longer than the vendors keep
// their task records (BytePlus: 7 days).
const VideoJobTTL = 8 * 24 * time.Hour

// Video job records are stored as opaque JSON; the proxyapi package owns
// their shape.
//
//	video_task:<sessionID>:<jobID>  provider side: a vendor task a session paid for
//	video_job:<jobID>               consumer side: a job this node submitted

func (s *SessionStorage) GetVideoTask(sessionID, jobID string) ([]byte, error) {
	return s.getOptional(formatKey("video_task", strings.ToLower(sessionID), jobID))
}

func (s *SessionStorage) PutVideoTask(sessionID, jobID string, record []byte) error {
	return s.db.SetWithTTL(formatKey("video_task", strings.ToLower(sessionID), jobID), record, VideoJobTTL)
}

// GetVideoTasks returns every task record of a session.
func (s *SessionStorage) GetVideoTasks(sessionID string) ([][]byte, error) {
	_, values, err := s.db.GetPrefixWithValues(formatPrefix("video_task", strings.ToLower(sessionID)))
	if err != nil {
		return nil, fmt.Errorf("error reading video tasks: %w", err)
	}
	return values, nil
}

func (s *SessionStorage) GetVideoJob(jobID string) ([]byte, error) {
	return s.getOptional(formatKey("video_job", jobID))
}

func (s *SessionStorage) PutVideoJob(jobID string, record []byte) error {
	return s.db.SetWithTTL(formatKey("video_job", jobID), record, VideoJobTTL)
}

func (s *SessionStorage) GetVideoJobs() ([][]byte, error) {
	_, values, err := s.db.GetPrefixWithValues(formatPrefix("video_job"))
	if err != nil {
		return nil, fmt.Errorf("error reading video jobs: %w", err)
	}
	return values, nil
}

// getOptional returns (nil, nil) for a missing key.
func (s *SessionStorage) getOptional(key []byte) ([]byte, error) {
	val, err := s.db.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading %s: %w", key, err)
	}
	return val, nil
}
