package PictureFS

import "io/fs"

type DirEntry fileStat

func (de *DirEntry) Name() string {
	return (*fileStat)(de).Name()
}

func (de *DirEntry) IsDir() bool {
	return (*fileStat)(de).IsDir()
}

func (de *DirEntry) Type() fs.FileMode {
	return (*fileStat)(de).Mode()
}

func (de *DirEntry) Info() (fs.FileInfo, error) {
	return (*fileStat)(de), nil
}
