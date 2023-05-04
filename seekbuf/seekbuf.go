// seekbuf implements a read-seekable buffer.
package seekbuf

import "io"

// Buffer is a ReadWriteCloser that supports seeking. It's intended to
// replicate the functionality of bytes.Buffer that I use in my projects.
//
// Note that the seeking is limited to the read marker; all writes are
// append-only.
type Buffer struct {
	data []byte
	pos  int
}

func New(data []byte) *Buffer {
	return &Buffer{
		data: data,
	}
}

func (b *Buffer) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}

	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

// Seek sets the read pointer to pos.
func (b *Buffer) Seek(pos int) {
	b.pos = pos
}

// Rewind resets the read pointer to 0.
func (b *Buffer) Rewind() {
	b.Seek(0)
}

// Close clears all the data out of the buffer and sets the read position to 0.
func (b *Buffer) Close() error {
	b.data = nil
	b.pos = 0
	return nil
}

// Len returns the length of data remaining to be read.
func (b *Buffer) Len() int {
	return len(b.data[b.pos:])
}
