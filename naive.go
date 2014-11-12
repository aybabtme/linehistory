package linehistory

import (
	"bytes"
)

type naive struct {
	maxSize int
	buffer  []byte

	// sep is used to find the begining of the first valid
	// line in the buffer, starting from tail and backward
	sep byte
}

// NewNaive creates a History that will never grow to use more than
// maxSize bytes of memory. The separator is used to determine
// the end of a line. The history is held in a naive buffer.
func NewNaive(maxSize int, sep byte) History {
	return &naive{
		maxSize: maxSize,
		buffer:  make([]byte, 0, maxSize),
		sep:     sep,
	}
}

func (r *naive) Len() int { return len(r.buffer) }

func (r *naive) Cap() int { return cap(r.buffer) }

// Add puts the bytes in the naive, discarding extra bytes
// if the naive was full. Each `sep` delimited parts of the
// bytes will be considered as a line.
//
// Adding a line that is longer than the max size of the naive
// will panic.
func (r *naive) Add(b []byte) {

	if len(b) > r.maxSize {
		b = b[len(b)-r.maxSize:]
	}

	if len(b)+len(r.buffer) <= r.maxSize {
		r.buffer = append(r.buffer, b...)
		return
	}

	// buffer too full to accept bytes

	// trim heading lines
	i := bytes.IndexByte(r.buffer, r.sep)
	if i >= 0 {
		r.buffer = r.buffer[i+1:]
	}

	// check if still too full
	if len(b)+len(r.buffer) <= r.maxSize {
		r.buffer = append(r.buffer, b...)
		return
	}

	avail := r.maxSize - len(r.buffer)
	need := len(b)
	toFree := need - avail
	if toFree > 0 {
		r.buffer = r.buffer[toFree:]
	}

	r.buffer = append(r.buffer, b...)
}

// Walk over the lines in the buffer, from the oldest to the
// newest.
func (r *naive) Walk(walk func([]byte)) {

	lastHead := 0

	for i := range r.buffer {
		if r.buffer[i] == r.sep {
			tail := i + 1
			walk(r.buffer[lastHead:tail])
			lastHead = tail
		}
	}
}
