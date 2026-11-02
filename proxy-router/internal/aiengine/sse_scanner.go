package aiengine

import (
	"bufio"
	"io"
)

// sseMaxTokenSize is the per-line cap for SSE frames. The default
// bufio.MaxScanTokenSize (64 KiB) is too small for large content, tool-call,
// or reasoning deltas — see #933.
const sseMaxTokenSize = 16 * 1024 * 1024

func newSSEScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), sseMaxTokenSize)
	return scanner
}
