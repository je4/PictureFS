package imagecollage

import (
	"golang.org/x/image/draw"
	"image"
	"os"
)

type Layout struct {
	Width, Height int64
	Rects         []Rect
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

func (layout Layout) CreateImage() (image.Image, error) {
	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: int(layout.Width), Y: int(layout.Height)}
	collImg := image.NewNRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
	//fmt.Printf("target: %v\n", collImg.Rect)
	for _, rect := range layout.Rects {
		draw.Copy(collImg)
	}
	return collImg, nil
}
