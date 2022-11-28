package encoding

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Reader interface {
	Reset()
	ResetAt(off int) error
	Read(int) []byte
	Peek(int) []byte
	Len() int

	// TODO: Seems like we need a "Remaining()" method here because Decbuf
	//  expects to be able to do `d.B = d.B[n:]` when it consumes parts of the
	//  underlying byte slice.

	// TODO: Add Seek method here to allow us to skip bytes without
	//  needing to read them or allocate? Easy to implement for BufReader
	//  and supported by bufio.Reader used in FileReader via Discard()
}

type BufReader struct {
	initial []byte
	b       []byte
}

func NewBufReader(bs ByteSlice) *BufReader {
	b := bs.Range(0, bs.Len())
	r := &BufReader{initial: b}
	r.Reset()
	return r
}

func (b *BufReader) Reset() {
	b.b = b.initial
}

func (b *BufReader) ResetAt(off int) error {
	b.b = b.initial
	if len(b.b) < off {
		return ErrInvalidSize
	}
	b.b = b.b[off:]
	return nil
}

func (b *BufReader) Peek(n int) []byte {
	if len(b.b) < n {
		n = len(b.b)
	}
	res := b.b[:n]
	return res
}

func (b *BufReader) Read(n int) []byte {
	if len(b.b) < n {
		n = len(b.b)
	}
	res := b.b[:n]
	b.b = b.b[n:]
	return res
}

func (b *BufReader) Len() int {
	return len(b.b)
}

type FileReader struct {
	file   *os.File
	buf    *bufio.Reader
	base   int
	length int
}

func NewFileReader(file *os.File, base, length int) *FileReader {
	f := &FileReader{
		file:   file,
		buf:    bufio.NewReader(file),
		base:   base,
		length: length,
	}
	f.Reset()
	return f
}

func (f *FileReader) Reset() {
	if _, err := f.file.Seek(int64(f.base), io.SeekStart); err != nil {
		fmt.Printf("seek: %v\n", err)
	}
	f.buf.Reset(f.file)
}

func (f *FileReader) ResetAt(off int) error {
	if _, err := f.file.Seek(int64(f.base+off), io.SeekStart); err != nil {
		fmt.Printf("seek: %v\n", err)
	}
	f.buf.Reset(f.file)
	return nil
}

func (f *FileReader) Peek(n int) []byte {
	b, err := f.buf.Peek(n)
	fmt.Printf("peek: %v\n", err)
	if len(b) > 0 {
		return b
	}
	return nil
}

func (f *FileReader) Read(n int) []byte {
	b := make([]byte, n)
	n, err := f.buf.Read(b)
	fmt.Printf("read: %v\n", err)
	if n > 0 {
		return b[:n]
	}
	return nil
}

func (f *FileReader) Len() int {
	// TODO: BufReader returns whatever is left in the backing byte slice
	//  here while _we_ only ever return the original size of the file (instead
	//  of how much of the file we haven't yet read). Is this a problem? Seems
	//  like it will be.
	return f.length
}
