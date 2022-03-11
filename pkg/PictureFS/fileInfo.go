package PictureFS

import (
	"io/fs"
	"time"
)

type FileMode = fs.FileMode
type FileInfo = fs.FileInfo

type fileStat struct {
	name string
	size int64
	dir  bool
}

func (fStat *fileStat) Name() string {
	return fStat.name
}

func (fStat *fileStat) isSymlink() bool {
	return false
}

func (fStat *fileStat) Size() int64 {
	return fStat.size
}

func (fStat *fileStat) Mode() (m FileMode) {
	if fStat.dir {
		return fs.ModeDir
	}
	return 0444
}

func (fStat *fileStat) ModTime() time.Time {
	return time.Unix(0, 0)
}

// Sys returns syscall.Win32FileAttributeData for file fs.
func (fStat *fileStat) Sys() interface{} {
	return nil
}

func (fStat *fileStat) IsDir() bool {
	return fStat.Mode().IsDir()
}
