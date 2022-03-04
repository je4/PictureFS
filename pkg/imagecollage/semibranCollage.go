package imagecollage

import (
	"fmt"
	"github.com/pkg/errors"
)

type SemibranCollage struct {
	rects []Rect
}

func NewSemibranCollage() *SemibranCollage {
	var sc = &SemibranCollage{rects: []Rect{}}
	return sc
}

func (sc *SemibranCollage) AddRect(name string, width, height int64) error {
	for _, r := range sc.rects {
		if r.Name == name {
			return errors.New(fmt.Sprintf("rectangle with name %s already added", name))
		}
	}
	sc.rects = append(sc.rects, Rect{
		Name:   name,
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
	})
	return nil
}
func (sc *SemibranCollage) Pack() (Layout, error) {
	layout := pack(sc.rects)
	return layout, nil
}
