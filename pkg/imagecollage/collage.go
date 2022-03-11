package imagecollage

import (
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"image"
)

type Rect struct {
	Name          string
	X, Y          int64
	Width, Height int64
}

type Collage interface {
	AddImageFile(path string) error
	AddRect(name string, width, height int64) error
	Pack() (Layout, error)
	CreateImage(layout Layout, dirName string) (image.Image, error)
	CreateLayout(layout Layout) (*PictureFS.Layout, error)
	CreateJSON(layout Layout) ([]byte, error)
}
