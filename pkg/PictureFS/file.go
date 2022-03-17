package PictureFS

import (
	"errors"
	"fmt"
	"io"
)

type File struct {
	name string
	fs   *FS
	i    int64
}

func (f *File) Close() error {
	f.i = 0

	return nil
}

func (f *File) Stat() (FileInfo, error) {
	if !f.fs.hasFile(f.name) {
		if !f.fs.hasDir(f.name) {
			return nil, errors.New(fmt.Sprintf("invalid node: %s", f.name))
		}
		return &fileStat{
			name: f.name,
			size: 0,
			dir:  true,
		}, nil
	}
	return &fileStat{
		name: f.name,
		size: int64(len(f.fs.data[f.name])), // hasFile makes sure, it exists,
		dir:  false,
	}, nil
}

// Len returns the number of bytes of the unread portion of the
// string.
func (f *File) Len() int64 {
	if !f.fs.hasFile(f.name) {
		return 0
	}
	var l int64 = int64(len(f.fs.data[f.name]))
	if f.i >= l {
		return 0
	}
	return l - f.i
}

func (f *File) Size() int64 {
	if !f.fs.hasFile(f.name) {
		return 0
	}
	return int64(len(f.fs.data[f.name]))
}

func (f *File) Read(buf []byte) (n int, err error) {
	if !f.fs.hasFile(f.name) {
		return 0, errors.New(fmt.Sprintf("invalid file: %s", f.name))
	}
	if f.i >= int64(len(f.fs.data[f.name])) {
		return 0, io.EOF
	}
	n = copy(buf, f.fs.data[f.name][f.i:])
	f.i += int64(n)
	return
}

// ReadAt implements the io.ReaderAt interface.
func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("PictureFS.File.ReadAt: negative offset")
	}
	if off >= int64(len(f.fs.data[f.name])) {
		return 0, io.EOF
	}
	n = copy(b, f.fs.data[f.name][off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

// ReadByte implements the io.ByteReader interface.
func (f *File) ReadByte() (byte, error) {
	if f.i >= int64(len(f.fs.data[f.name])) {
		return 0, io.EOF
	}
	b := f.fs.data[f.name][f.i]
	f.i++
	return b, nil
}

// Seek implements the io.Seeker interface.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.i + offset
	case io.SeekEnd:
		abs = int64(len(f.fs.data[f.name])) + offset
	default:
		return 0, errors.New("strings.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("PictureFS.File.Seek: negative position")
	}
	f.i = abs
	return abs, nil
}
