package imagecollage

type Rect struct {
	Name          string
	X, Y          int64
	Width, Height int64
}

type Collage interface {
	AddRect(name string, width, height int64) error
	Pack() (Layout, error)
}
