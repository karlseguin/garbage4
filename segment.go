package garbage4

import (
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const SEGMENT_HEADER_SIZE = 28

type Segment struct {
	header *SegmentHeader
	ref    []byte
	file   *os.File
	data   *[MAX_QUEUE_SIZE]byte
}

type SegmentHeader struct {
	version uint32
	flag    uint32
	size    uint32
	id      uint64
	nextId  uint64
}

func newSegment(t *Topic) *Segment {
	id := uint64(time.Now().UnixNano())
	segment := openSegment(t, id, true)
	segment.header.id = id
	return segment
}

func openSegment(t *Topic, id uint64, truncate bool) *Segment {
	name := PATH + t.name + "/" + strconv.FormatUint(id, 10) + ".q"
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if truncate {
		file.Truncate(MAX_QUEUE_SIZE)
	}
	ref, err := syscall.Mmap(int(file.Fd()), 0, MAX_QUEUE_SIZE, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}

	s := &Segment{
		ref:  ref,
		file: file,
		data: (*[MAX_QUEUE_SIZE]byte)(unsafe.Pointer(&ref[0])),
	}
	s.header = (*SegmentHeader)(unsafe.Pointer(&s.data[0]))
	s.header.size = SEGMENT_HEADER_SIZE
	return s
}

func (s *Segment) Close() {
	syscall.Munmap(s.ref)
	s.file.Close()
	s.data, s.ref = nil, nil
}
