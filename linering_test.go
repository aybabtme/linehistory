package linehistory_test

import (
	"bytes"
	crand "crypto/rand"
	"github.com/aybabtme/linehistory"
	mrand "math/rand"
	"testing"
)

func TestCanAddUniqueLine(t *testing.T) {
	ring := linehistory.NewRing(20, '\n')

	msg := []byte("hello!\n")
	ring.Add(msg)

	var got [][]byte
	ring.Walk(func(b []byte) {
		got = append(got, b)
	})

	if len(got) != 1 {
		t.Fatalf("want len 1, got %d", len(got))
	}

	if !bytes.Equal(got[0], msg) {
		t.Fatalf("want %x got %x", msg, got[0])
	}
}

func TestCanAddTwoLines(t *testing.T) {
	ring := linehistory.NewRing(20, '\n')

	want := [][]byte{
		[]byte("hello!\n"),
		[]byte("bye!\n"),
	}

	for _, l := range want {
		ring.Add(l)
	}

	var got [][]byte
	ring.Walk(func(b []byte) {
		got = append(got, b)
	})

	if len(got) != len(want) {
		t.Fatalf("want len %d, got %d", len(want), len(got))
	}

	for i, gotLine := range got {
		wantLine := want[i]
		t.Logf("idx =%d", i)
		t.Logf("want=%x", wantLine)
		t.Logf("got =%x", gotLine)
		if !bytes.Equal(wantLine, gotLine) {
			t.Errorf("mismatch at index %d!", i)
		}
	}

	t.Logf("\n%v", ring)
}

func TestCanOverflowRing(t *testing.T) {
	ring := linehistory.NewRing(20, '\n')

	input := [][]byte{
		[]byte("hello!\n"), // this line will be evicted
		[]byte("derp derp!\n"),
		[]byte("bye!\n"),
	}

	want := [][]byte{
		[]byte("derp derp!\n"),
		[]byte("bye!\n"),
	}

	for _, l := range input {
		ring.Add(l)
		t.Logf("len=%d, cap=%d", ring.Len(), ring.Cap())
	}

	var got [][]byte
	ring.Walk(func(b []byte) {
		got = append(got, b)
	})

	if len(got) != len(want) {
		t.Errorf("want len %d, got %d", len(want), len(got))
	}

	for i, gotLine := range got {
		wantLine := want[i]
		t.Logf("idx =%d", i)
		t.Logf("want=%x (%q)", wantLine, string(wantLine))
		t.Logf("got =%x (%q)", gotLine, string(gotLine))
		if !bytes.Equal(wantLine, gotLine) {
			t.Errorf("mismatch at index %d!", i)
		}
	}

	t.Logf("\n%v", ring)
}

func TestAddLineBiggerThanBufferPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic, got nothing")
		}
	}()

	ring := linehistory.NewRing(10, '\n')
	ring.Add([]byte("i am more than 10 bytes long\n"))
}

func TestAddingLineSameSize(t *testing.T) {
	length := 20
	ring := linehistory.NewRing(length, '\n')
	data := append(bytes.Repeat([]byte{0xFF}, length-1), '\n')

	ring.Add(data)

	var gotLines [][]byte
	ring.Walk(func(b []byte) {
		gotLines = append(gotLines, b)
	})

	if len(gotLines) != 1 {
		t.Errorf("want 1 line, got %d", len(gotLines))
	}

	t.Logf("%v", ring)
}

func TestAddingLineSameSizeManyTimes(t *testing.T) {
	length := 20
	ring := linehistory.NewRing(length, '\n')
	data := append(bytes.Repeat([]byte{0xFF}, length-1), '\n')

	for i := 0; i < 100; i++ {
		ring.Add(data)
	}

	var gotLines [][]byte
	ring.Walk(func(b []byte) {
		gotLines = append(gotLines, b)
	})

	if len(gotLines) != 1 {
		t.Errorf("want 1 line, got %d", len(gotLines))
	}

	t.Logf("%v", ring)
}

func TestAddingLineFitsOnlyOneManyTimes(t *testing.T) {

	data := []byte("herp\n")
	ring := linehistory.NewRing(len(data)+2, '\n')

	for i := 0; i < 1000; i++ {
		ring.Add(data)

		var gotLines [][]byte
		ring.Walk(func(b []byte) {
			gotLines = append(gotLines, b)
		})

		if len(gotLines) != 1 {
			t.Errorf("want 1 line, got %d", len(gotLines))
		}

	}

	t.Logf("%v", ring)
}

func TestNeverExceedMaxSize(t *testing.T) {
	length := 23
	ring := linehistory.NewRing(length, '\n')
	for i := 0; i < 100000; i++ {
		ring.Add(randByte(mrand.Intn(length-1) + 1))
		if ring.Cap() > length {
			t.Fatalf("want capacity of %d, got %d", length, ring.Cap())
		}
	}
}

func randByte(n int) []byte {
	b := make([]byte, n-1)
	_, _ = crand.Read(b)
	return append(b, '\n')
}
