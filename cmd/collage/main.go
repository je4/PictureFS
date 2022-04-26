package main

import (
	"flag"
	"fmt"
	"github.com/je4/PictureFS/v2/pkg/imagecollage"
	"github.com/pkg/errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const borderWidth = 3
const offsetX = 20
const offsetY = 20

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open image %s", filePath)
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode image %s", filePath)
	}

	return image, nil
}

func main() {
	var basedir = flag.String("folder", ".", "base folder with image contents")
	var marginExt = flag.Int64("margin", 20, "empty margin around collage")
	var border = flag.Int64("border", 2, "width of black border around each image")
	var space = flag.Int64("space", 2, "empty space around images")
	var output = flag.String("output", "./collage.png", "name of output image (metadata json file is same with extension .json")

	flag.Parse()

	var folder = filepath.ToSlash(filepath.Clean(*basedir))

	var collage imagecollage.Collage

	collage = imagecollage.NewSemibranCollage(
		*basedir,
		*border,
		*space,
		*marginExt,
		*marginExt,
		*marginExt,
		*marginExt)

	filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		imgPath := strings.TrimPrefix(filepath.ToSlash(path), folder)
		if err := collage.AddImageFile(imgPath); err != nil {
			log.Printf("%s not an image: %v", imgPath, err)
			//			return errors.Wrapf(err, "cannot add image %s", path)
		} else {
			log.Printf("adding image %s", imgPath)
		}
		return nil
	})

	layout, err := collage.Pack()
	if err != nil {
		log.Fatalf("cannot pack: %v", err)
	}
	for _, rect := range layout.Rects {
		fmt.Printf("%v\n", rect)
	}
	result, err := collage.CreateImage(layout, folder)
	if err != nil {
		log.Fatalf("cannot create target image: %v", err)
	}
	outimg := filepath.Clean(*output)
	fDst, err := os.Create(outimg)
	if err != nil {
		log.Fatal(err)
	}
	defer fDst.Close()
	err = png.Encode(fDst, result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("output image written: %s\n", outimg)

	outjson := filepath.Clean(*output) + ".json"
	jsonBytes, err := collage.CreateJSON(layout)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outjson, jsonBytes, 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("output json written: %s\n", outjson)

}
