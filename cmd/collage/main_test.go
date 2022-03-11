package main

import (
	"fmt"
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"github.com/je4/PictureFS/v2/pkg/imagecollage"
	"image"
	"image/color"
	"image/png"
	"io/fs"
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

	collage = imagecollage.NewSemibranCollage(
		dname,
		2,
		2,
		10,
		20,
		10,
		10)

	for i := 0; i < 50; i++ {
		width := 2 * (rand.Intn(MAX_W) + 1)
		height := 2 * (rand.Intn(MAX_H) + 1)
		fname := fmt.Sprintf("pictureFS_%04dx%04d.png", width, height)
		dir := fmt.Sprintf("%v", i%5)
		dir2 := filepath.Join(dname, dir)
		if err := os.MkdirAll(dir2, 0777); err != nil {
			t.Fatalf("cannot create dir %s", dir2)
		}
		fullpath := filepath.Join(dir2, fname)
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
		collage.AddImageFile(filepath.Join(dir, fname))

		fmt.Printf("Image %s\n", fullpath)
	}
	layout, err := collage.Pack()
	if err != nil {
		t.Fatalf("cannot pack: %v", err)
	}
	for _, rect := range layout.Rects {
		fmt.Printf("%v\n", rect)
	}
	result, err := collage.CreateImage(layout, dname)
	if err != nil {
		t.Fatalf("cannot create target image: %v", err)
	}
	outimg := filepath.Join(dname, "out.png")
	fDst, err := os.Create(outimg)
	if err != nil {
		t.Fatal(err)
	}
	defer fDst.Close()
	err = png.Encode(fDst, result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("output image written: %s\n", outimg)

	outjson := filepath.Join(dname, "out.json")
	jsonBytes, err := collage.CreateJSON(layout)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outjson, jsonBytes, 0666); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("output json written: %s\n", outjson)
	for _, rect := range layout.Rects {
		if err := os.Remove(filepath.Join(dname, rect.Name)); err != nil {
			t.Fatalf("cannot remove file %s: %v", filepath.Join(dname, rect.Name), err)
		}
	}

	pfs, err := PictureFS.NewFSFile(outimg, outjson)
	if err != nil {
		t.Fatalf("cannot create picture fs %s/%s", outimg, outjson)
	}

	if err := PictureFS.WalkDir(pfs, "/", func(path string, d fs.DirEntry, err error) error {
		fmt.Printf("%s: dir: %v\n", path, d.IsDir())
		return nil
	}); err != nil {
		t.Fatalf("cannot walk directory: %v", err)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
