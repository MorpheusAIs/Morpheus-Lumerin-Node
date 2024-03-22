package interfaces

import (
	"io"
	"sync"
)

const (
	KiloByte             = 1024
	MegaByte             = 1024 * KiloByte
	DefaultCapacityBytes = 1 * MegaByte
)

type LogStorage struct {
	Buffer MyBuf
	id     string
}

func NewLogStorage(ID string) *LogStorage {
	return &LogStorage{
		Buffer: NewMyBuf(DefaultCapacityBytes),
		id:     ID,
	}
}

func NewLogStorageWithCapacity(ID string, capBytes uint64) *LogStorage {
	return &LogStorage{
		Buffer: NewMyBuf(capBytes),
		id:     ID,
	}
}

func (l *LogStorage) ID() string {
	return l.id
}

func (l *LogStorage) GetReader() io.Reader {
	return l.Buffer.NewReader()
}

type MyBuf interface {
	io.Writer
	NewReader() io.Reader
	Capacity() uint64
}

type mybuf struct {
	data [][]byte
	len  uint64
	cap  uint64
	sync.RWMutex
}

func NewMyBuf(cap uint64) MyBuf {
	return &mybuf{
		cap: cap,
	}
}

func (mb *mybuf) Capacity() uint64 {
	return mb.cap
}

func (mb *mybuf) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Cannot retain p, so we must copy it:
	p2 := make([]byte, len(p))
	copy(p2, p)
	mb.Lock()
	for {
		newLen := mb.len + uint64(len(p2))
		if newLen >= mb.cap {
			if len(mb.data) == 0 {
				break
			}
			mb.len -= uint64(len(mb.data[0]))
			mb.data = mb.data[1:]
		} else {
			mb.len = newLen
			break
		}
	}
	if len(p2) > int(mb.cap) {
		p2 = p2[len(p2)-int(mb.cap):]
	}

	mb.data = append(mb.data, p2)
	mb.Unlock()
	return len(p), nil
}

type mybufReader struct {
	mb   *mybuf // buffer we read from
	i    int    // next slice index
	data []byte // current data slice to serve
}

func (mbr *mybufReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Do we have data to send?
	if len(mbr.data) == 0 {
		mb := mbr.mb
		mb.RLock()
		if mbr.i < len(mb.data) {
			mbr.data = mb.data[mbr.i]
			mbr.i++
		}
		mb.RUnlock()
	}
	if len(mbr.data) == 0 {
		return 0, io.EOF
	}

	n = copy(p, mbr.data)
	mbr.data = mbr.data[n:]
	return n, nil
}

func (mb *mybuf) NewReader() io.Reader {
	return &mybufReader{mb: mb}
}
