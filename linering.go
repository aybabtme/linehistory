package linehistory

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
)

// History of bytes, line oriented. Will keep an history of
// the previous lines it received, and discard the oldest
// ones when it runs out of space.
type History interface {
	Add([]byte)
	Walk(walk func([]byte))
	Len() int
	Cap() int
}

type ring struct {
	buffer []byte

	head int
	tail int

	// sep is used to find the begining of the first valid
	// line in the buffer, starting from tail and backward
	sep byte
}

// NewRing creates a History that will never grow to use more than
// maxSize bytes of memory. The separator is used to determine
// the end of a line. The history is held in a ring buffer.
func NewRing(maxSize int, sep byte) History {
	return &ring{
		buffer: make([]byte, maxSize),
		head:   0,
		tail:   0,
		sep:    sep,
	}
}

func (r *ring) Len() int {
	if r.head < r.tail {
		return r.tail - r.head
	}
	return (r.tail + len(r.buffer)) - r.head
}

func (r *ring) Cap() int { return cap(r.buffer) }

// Add puts the bytes in the ring, discarding extra bytes
// if the ring was full. Each `sep` delimited parts of the
// bytes will be considered as a line.
//
// Adding a line that is longer than the max size of the ring
// will panic.
func (r *ring) Add(b []byte) {
	// if head beyond tail, need to wrap over
	newTail := (len(b) + r.tail)

	if newTail > len(r.buffer) {
		newTail %= len(r.buffer)

		overflow := len(r.buffer) - r.tail
		copy(r.buffer[r.tail:len(r.buffer)], b[:overflow])
		copy(r.buffer[0:newTail], b[overflow:])

		// advance head to one byte past next r.sep in buffer
		index := bytes.IndexByte(r.buffer[newTail+1:r.tail], r.sep)
		if index == -1 {
			r.head = r.tail
		} else {
			realIdx := newTail + 1 + index
			r.head = (realIdx + 1) % len(r.buffer)
		}
	} else {
		copy(r.buffer[r.tail:newTail], b)
	}
	r.tail = newTail
}

// Walk over the lines in the buffer, from the oldest to the
// newest.
func (r *ring) Walk(walk func([]byte)) {

	bytesBetween := r.Len()
	lastHead := r.head
	log.Printf("walk lastHead=%d\tbytesBetween=%d", lastHead, bytesBetween)

	for i := r.head; i < r.head+bytesBetween; i++ {
		idx := imin(i%len(r.buffer), i)

		if r.buffer[idx] == r.sep {
			tail := idx + 1
			if tail < lastHead {
				walk(append(
					r.buffer[lastHead:],
					r.buffer[:tail]...,
				))
			} else {
				walk(r.buffer[lastHead:tail])
			}

			lastHead = tail
		}
		log.Printf("idx=%d", idx)
	}
}

func (r *ring) String() string {
	data := hex.Dump(r.buffer)
	return data + fmt.Sprintf("head=%d\ntail=%d", r.head, r.tail)
}

func imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
