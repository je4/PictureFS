package PictureFS

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Sub(fsys fs.FS, dir string) (fs.FS, error) {
	lfs, ok := fsys.(*FS)
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid filesystem type: %T", fsys))
	}
	newDir := filepath.ToSlash(filepath.Clean(filepath.Join(lfs.base, dir)))
	if !lfs.hasDir(newDir) {
		return nil, errors.New(fmt.Sprintf("invalid directory: %s", newDir))
	}
	subFS := &FS{
		base: newDir,
		data: lfs.data,
	}
	return subFS, nil
}

func FileInfoToDirEntry(info fs.FileInfo) fs.DirEntry {
	if info == nil {
		return nil
	}
	fi, ok := info.(*fileStat)
	if !ok {
		log.Fatalf("invalid fileinfo %T not fileStat", info)
	}
	return (*DirEntry)(fi)
}

func ReadFile(fsys fs.FS, name string) ([]byte, error) {
	pfs, ok := fsys.(*FS)
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid filesystem type %T", fsys))
	}
	f, err := pfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func ValidPath(name string) bool {
	path := filepath.ToSlash(name)
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if p == ".." {
			return false
		}
	}
	return true
}

type fsData map[string][]byte

func (pfs *FS) dirEntries(dir string) []string {
	result := []string{}
	dir = strings.TrimRight(dir, "/") + "/"
	for d, _ := range pfs.data {
		d = filepath.ToSlash(filepath.Join(pfs.base, d))
		if strings.HasPrefix(d, dir) {
			h := strings.TrimPrefix(d, dir)
			parts := strings.Split(h, "/")
			if len(parts) > 0 {
				fp := filepath.ToSlash(filepath.Clean(filepath.Join(dir, parts[0])))
				found := false
				for _, h := range result {
					if h == fp {
						found = true
						break
					}
				}
				if !found {
					result = append(result, fp)
				}
			}
		}
	}
	return result
}

func (pfs *FS) hasDir(dir string) bool {
	//dir = filepath.ToSlash(filepath.Clean(dir))
	dir = strings.TrimRight(dir, "/") + "/"
	for d, _ := range pfs.data {
		d2 := filepath.ToSlash(filepath.Join(pfs.base, d))
		if strings.HasPrefix(d2, dir) && len(d2) > len(dir) {
			return true
		}
	}
	return false
}

func (pfs *FS) hasFile(path string) bool {
	path = filepath.ToSlash(filepath.Clean(path))

	_, ok := pfs.data[path]
	return ok
}

type FS struct {
	base string
	data fsData
}

func NewFSFile(img string, layout string) (*FS, error) {
	fImg, err := os.Open(img)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open image file %s", img)
	}
	defer fImg.Close()
	image, _, err := image.Decode(fImg)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode image file %s", img)
	}
	fJSON, err := os.Open(layout)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open json file %s", layout)
	}
	defer fJSON.Close()
	dec := json.NewDecoder(fJSON)
	l := Layout{}
	if err := dec.Decode(&l); err != nil {
		return nil, errors.Wrapf(err, "cannot decode json file %s", layout)
	}
	return NewFS(image, l)
}

func NewFS(img image.Image, layout Layout) (*FS, error) {
	pfs := &FS{
		base: "/",
		data: make(fsData),
	}
	for _, rect := range layout.Images {
		newImg := image.NewNRGBA(image.Rectangle{
			Min: image.Point{},
			Max: image.Point{X: rect.Width, Y: rect.Height},
		})
		draw.Copy(newImg,
			image.Point{},
			img,
			image.Rectangle{
				Min: image.Point{X: rect.X, Y: rect.Y},
				Max: image.Point{X: rect.X + rect.Width, Y: rect.Y + rect.Height},
			},
			draw.Over,
			nil,
		)
		var data = bytes.NewBuffer(nil)
		var err error
		ext := strings.ToLower(filepath.Ext(rect.Path))
		switch ext {
		case ".jpg":
			err = jpeg.Encode(data, newImg, nil)
		case ".jpeg":
			err = jpeg.Encode(data, newImg, nil)
			//		case ".png":
			//			err = png.Encode(data, newImg)
		case ".gif":
			err = gif.Encode(data, newImg, nil)
		default:
			err = png.Encode(data, newImg)
			//return nil, errors.New(fmt.Sprintf("invalid image extension %s in path %s", ext, rect.Path))
		}
		if err != nil {
			return nil, errors.Wrapf(err, "cannot encode image %s", rect.Path)
		}
		pfs.data[strings.Replace(
			filepath.ToSlash(
				filepath.Clean(
					"/"+strings.TrimPrefix(rect.Path, "/"))), "//", "/", -1)] = data.Bytes()
	}
	return pfs, nil
}

func (pfs *FS) Open(name string) (fs.File, error) {
	fullpath := "/" + strings.TrimPrefix(filepath.ToSlash(filepath.Clean(filepath.Join(pfs.base, name))), "/")
	if !pfs.hasFile(fullpath) {
		return nil, fs.ErrNotExist
	}
	return &File{
		name: fullpath,
		fs:   pfs,
		i:    0,
	}, nil
}

func (pfs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = "/" + strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "/")
	// check could be removed, but error message is better than just empty result
	if !pfs.hasDir(name) {
		return nil, errors.New(fmt.Sprintf("%s is not a directory", name))
	}
	entries := pfs.dirEntries(name)
	// sort on filename
	sort.Strings(entries)
	dEntries := []fs.DirEntry{}
	for _, p := range entries {
		f := &File{
			name: p,
			fs:   pfs,
			i:    0,
		}
		fi, err := f.Stat()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot stat file %s", p)
		}
		dEntries = append(dEntries, FileInfoToDirEntry(fi))
	}
	return dEntries, nil
}

func WalkDir(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
	root = "/" + strings.TrimPrefix(filepath.ToSlash(filepath.Clean(root)), "/")
	pfs, ok := fsys.(*FS)
	if !ok {
		return errors.New(fmt.Sprintf("invalid filesystem type %T", fsys))
	}

	dirEntries, err := pfs.ReadDir(root)
	if err != nil {
		return errors.Wrapf(err, "cannot read %s", root)
	}
	for _, dirEntry := range dirEntries {
		subdir := filepath.ToSlash(filepath.Join(root, dirEntry.Name()))
		if err := fn(subdir, dirEntry, nil); err != nil {
			return err
		}
		if dirEntry.IsDir() {
			if err := WalkDir(fsys, subdir, fn); err != nil {
				return err
			}
		}
	}
	return nil
}
