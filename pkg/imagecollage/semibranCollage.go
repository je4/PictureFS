package imagecollage

import (
	"encoding/json"
	"fmt"
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"github.com/pkg/errors"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"os"
	"path/filepath"
)

type SemibranCollage struct {
	rects                                            []Rect
	basePath                                         string
	border                                           int64
	margin                                           int64
	marginTop, marginLeft, marginBottom, marginRight int64
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func DrawRect(x1, y1, x2, y2, thickness int, col color.Color, img *image.NRGBA) {

	for t := 0; t < thickness; t++ {
		// draw horizontal lines
		for x := x1; x <= x2; x++ {
			img.Set(x, y1+t, col)
			img.Set(x, y2-t, col)
		}
		// draw vertical lines
		for y := y1; y <= y2; y++ {
			img.Set(x1+t, y, col)
			img.Set(x2-t, y, col)
		}
	}
}

func NewSemibranCollage(
	basePath string,
	borderWidth, margin int64,
	marginLeft, marginTop, marginRight, marginBottom int64,
) *SemibranCollage {
	var sc = &SemibranCollage{
		rects:        []Rect{},
		basePath:     basePath,
		border:       borderWidth,
		margin:       margin,
		marginBottom: marginBottom,
		marginLeft:   marginLeft,
		marginRight:  marginRight,
		marginTop:    marginTop,
	}
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
		Width:  width + 2*sc.border + 2*sc.margin,
		Height: height + 2*sc.border + 2*sc.margin,
	})
	return nil
}

func (sc *SemibranCollage) AddImageFile(path string) error {
	path = filepath.ToSlash(filepath.Clean(path))
	fullpath := filepath.Join(sc.basePath, path)
	img, err := getImageFromFilePath(fullpath)
	if err != nil {
		return errors.Wrapf(err, "cannot open image %s", fullpath)
	}
	return sc.AddRect(path, int64(img.Bounds().Dx()), int64(img.Bounds().Dy()))
}

func (sc *SemibranCollage) Pack() (Layout, error) {
	layout := pack(sc.rects)
	return layout, nil
}

func (sc *SemibranCollage) CreateImage(layout Layout, dirName string) (image.Image, error) {
	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{
		X: int(layout.Width + sc.marginLeft + sc.marginRight),
		Y: int(layout.Height + sc.marginTop + sc.marginBottom),
	}
	collImg := image.NewNRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
	//fmt.Printf("target: %v\n", collImg.Rect)
	for _, rect := range layout.Rects {
		src, err := getImageFromFilePath(filepath.Join(dirName, rect.Name))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot open image %s", rect.Name)
		}
		if sc.border > 0 {
			DrawRect(
				int(rect.X+sc.marginLeft+sc.margin),
				int(rect.Y+sc.marginTop+sc.margin),
				int(rect.X+sc.marginLeft+sc.margin+int64(src.Bounds().Dx())+2*sc.border-1),
				int(rect.Y+sc.marginTop+sc.margin+int64(src.Bounds().Dy())+2*sc.border-1),
				int(sc.border), color.Black, collImg)
		}
		draw.Copy(
			collImg,
			image.Point{
				X: int(rect.X + sc.marginLeft + sc.margin + sc.border),
				Y: int(rect.Y + sc.marginTop + sc.margin + sc.border),
			},
			src,
			src.Bounds(),
			draw.Over,
			nil,
		)
	}
	return collImg, nil
}

func (sc *SemibranCollage) CreateLayout(layout Layout) (*PictureFS.Layout, error) {
	var result = &PictureFS.Layout{
		Version: PictureFS.VERSION,
		Images:  []PictureFS.Rect{},
	}

	for _, rect := range layout.Rects {
		result.Images = append(result.Images, PictureFS.Rect{
			Path:   rect.Name,
			X:      int(rect.X + sc.border + sc.margin),
			Y:      int(rect.Y + sc.border + sc.margin),
			Width:  int(rect.Width - 2*sc.border - 2*sc.margin),
			Height: int(rect.Height - 2*sc.border - 2*sc.margin),
		})
	}

	return result, nil
}

func (sc *SemibranCollage) CreateJSON(layout Layout) ([]byte, error) {
	result, err := sc.CreateLayout(layout)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal json of %v", result)
	}
	return jsonBytes, nil
}
