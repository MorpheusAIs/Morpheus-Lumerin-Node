package storages

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type StorageLogger struct {
	log lib.ILogger
}

func NewStorageLogger(log lib.ILogger) *StorageLogger {
	return &StorageLogger{
		log: log.Named("Storage"),
	}
}

func (l *StorageLogger) Errorf(s string, p ...interface{}) {
	l.log.Errorf(normalize(s), p...)
}
func (l *StorageLogger) Warningf(s string, p ...interface{}) {
	l.log.Warnf(normalize(s), p...)
}
func (l *StorageLogger) Infof(s string, p ...interface{}) {
	l.log.Infof(normalize(s), p...)
}
func (l *StorageLogger) Debugf(s string, p ...interface{}) {
	l.log.Debugf(normalize(s), p...)
}

func normalize(s string) string {
	return strings.TrimRight(s, "\n")
	// trims new line
	// return strings.
}
