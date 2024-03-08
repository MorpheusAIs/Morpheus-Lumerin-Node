package interfaces

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogOverwrite(t *testing.T) {
	// Create a new log storage with a capacity of 1 byte.
	logStorage := NewLogStorageWithCapacity("test", 1)
	// Write a byte to the log.
	_, _ = logStorage.Buffer.Write([]byte{0})
	// Write another byte to the log.
	_, _ = logStorage.Buffer.Write([]byte{1})
	// Read the log.
	logData := make([]byte, 2)
	_, _ = logStorage.GetReader().Read(logData)
	// Check that the log contains the last byte written.
	if logData[0] != 1 {
		t.Errorf("Expected log to contain 1, got %d", logData[0])
	}
}

func TestLogOverwriteMessageLargerThanCap(t *testing.T) {
	// Create a new log storage with a capacity of 1 byte.
	logStorage := NewLogStorageWithCapacity("test", 1)
	// Write a byte to the log.
	_, _ = logStorage.Buffer.Write([]byte{0})
	// Write another byte to the log.
	_, _ = logStorage.Buffer.Write([]byte{1, 2})
	// Read the log.
	logData := make([]byte, 2)
	_, _ = logStorage.GetReader().Read(logData)
	// Check that the log contains the last byte written.
	if logData[0] != 2 {
		t.Errorf("Expected log to contain 2, got %d", logData[0])
	}
}

func TestLogOverwriteMessageVariableSize(t *testing.T) {
	logStorage := NewLogStorageWithCapacity("test", 5)
	_, _ = logStorage.Buffer.Write([]byte{1, 2})
	_, _ = logStorage.Buffer.Write([]byte{3, 4, 5})
	_, _ = logStorage.Buffer.Write([]byte{6})

	reader := logStorage.GetReader()

	logData := make([]byte, 4, 4)
	_, _ = reader.Read(logData)
	require.ElementsMatch(t, []byte{3, 4, 5, 0}, logData)

	logData = make([]byte, 4, 4)
	_, _ = reader.Read(logData)
	require.ElementsMatch(t, []byte{6, 0, 0, 0}, logData)

	logData = make([]byte, 4, 4)
	_, _ = reader.Read(logData)
	require.ElementsMatch(t, []byte{0, 0, 0, 0}, logData)
}

func TestLogOverwriteMessageVariableSize2(t *testing.T) {
	logStorage := NewLogStorageWithCapacity("test", 2)
	_, _ = logStorage.Buffer.Write([]byte{1})
	_, _ = logStorage.Buffer.Write([]byte{2})
	_, _ = logStorage.Buffer.Write([]byte{3})
	_, _ = logStorage.Buffer.Write([]byte{4, 5})

	reader := logStorage.GetReader()

	logData := make([]byte, 2, 2)
	_, _ = reader.Read(logData)
	require.ElementsMatch(t, []byte{4, 5}, logData)

	logData = make([]byte, 2, 2)
	_, _ = reader.Read(logData)
	require.ElementsMatch(t, []byte{0, 0}, logData)
}
