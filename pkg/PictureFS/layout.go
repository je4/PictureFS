package PictureFS

const VERSION = "0.1"

type Rect struct {
	Path          string
	X, Y          int
	Width, Height int
}

type Layout struct {
	Version string
	Images  []Rect
}
