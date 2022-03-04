package main

import (
	"fmt"
	"github.com/je4/PictureFS/v2/pkg/imagecollage"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const MAX_W = 100
const MAX_H = 100

func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

// Bresenham's algorithm, http://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
// TODO: handle int overflow etc.
func drawline(x0, y0, x1, y1 int, col color.Color, img *image.NRGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func TestCollage(t *testing.T) {
	dname, err := os.MkdirTemp("", "pictureFS")
	if err != nil {
		t.Fatalf("cannot create temp dir \"pictureFS\": %v", err)
	}
	fmt.Printf("Folder: %s\n", dname)
	//defer os.RemoveAll(dname)

	var collage imagecollage.Collage

	collage = imagecollage.NewSemibranCollage()

	for i := 0; i < 3; i++ {
		width := 2 * (rand.Intn(MAX_W) + 1)
		height := 2 * (rand.Intn(MAX_H) + 1)
		fname := fmt.Sprintf("pictureFS_%04dx%04d.png", width, height)
		fullpath := filepath.Join(dname, fname)
		f, err := os.Create(fullpath)
		if err != nil {
			t.Fatalf("cannot create temp dir \"pictureFS\": %v", err)
		}
		upLeft := image.Point{X: 0, Y: 0}
		lowRight := image.Point{X: width, Y: height}
		img := image.NewNRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
		Rect(0, 0, width-1, height-1, 2, color.RGBA{R: 0, G: 255, B: 0, A: 255}, img)
		drawline(0, 0, width-1, height-1, color.RGBA{R: 0, G: 255, B: 0, A: 255}, img)
		drawline(width-1, 0, 0, height-1, color.RGBA{R: 0, G: 255, B: 0, A: 255}, img)
		png.Encode(f, img)
		f.Close()
		collage.AddRect(fullpath, int64(width), int64(height))

		fmt.Printf("Image %s\n", fullpath)
	}
	layout, err := collage.Pack()
	if err != nil {
		t.Fatalf("cannot pack: %v", err)
	}
	for _, rect := range layout.Rects {
		fmt.Printf("%v\n", rect)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
