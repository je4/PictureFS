package imagecollage

import (
	"math"
	"sort"
)

// https://github.com/semibran/pack

// weights: greater side length produces more square-like output
const WHITESPACE_WEIGHT = 1
const SIDE_LENGTH_WEIGHT = 20

type Position struct {
	x, y int64
}

type Size struct {
	width, height int64
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// determines if rect `a` and rect `b` intersect
func intersects(a, b Rect) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// determines if the region specified by `rect` is clear of all other `Rects`
func validate(rects []Rect, rect Rect) bool {
	var a = rect
	for i := 0; i < len(rects); i++ {
		var b = rects[i]
		if intersects(a, b) {
			return false
		}
	}
	return true
}

// determines the amount of whitespace area remaining in `layout`
func whitespace(layout Layout) int64 {
	var whitespace = layout.Width * layout.Height
	for i := 0; i < len(layout.Rects); i++ {
		var rect = layout.Rects[i]
		whitespace -= rect.Width * rect.Height
	}
	return whitespace
}

// determine the desirability of a given layout
func rate(layout Layout) int64 {
	return whitespace(layout)*WHITESPACE_WEIGHT + max(layout.Width, layout.Height)*SIDE_LENGTH_WEIGHT
}

// finds the smallest `[ Width, Height ]` tuple that contains all `Rects`
func findBounds(rects []Rect) Size {
	var width, height int64
	for i := 0; i < len(rects); i++ {
		var rect = rects[i]
		var right = rect.X + rect.Width
		var bottom = rect.Y + rect.Height
		if right > width {
			width = right
		}
		if bottom > height {
			height = bottom
		}
	}
	return Size{
		width:  width,
		height: height,
	}
}

// find all rect positions given a rect list
func findPositions(rects []Rect) []Position {
	var positions = []Position{}
	for i := 0; i < len(rects); i++ {
		var rect = rects[i]
		for x := int64(0); x < rect.Width; x++ {
			positions = append(positions, Position{
				x: rect.X + x,
				y: rect.Y + rect.Height,
			})
		}
		for y := int64(0); y < rect.Height; y++ {
			positions = append(positions, Position{
				x: rect.X + rect.Width,
				y: rect.Y + y,
			})
		}
	}
	return positions
}

// finds the best location for a { Width, Height } tuple within the given layout
func findBestRect(layout Layout, size Rect) Rect {
	var bestRect = Rect{
		X:      0,
		Y:      0,
		Width:  size.Width,
		Height: size.Height,
	}

	if len(layout.Rects) <= 0 {
		return bestRect
	}

	var rect = Rect{
		X:      0,
		Y:      0,
		Width:  size.Width,
		Height: size.Height,
	}

	var sandbox = Layout{
		Width:  0,
		Height: 0,
		Rects:  layout.Rects,
	}

	var bestScore int64 = math.MaxInt64
	var positions = findPositions(layout.Rects)
	for i := 0; i < len(positions); i++ {
		var pos = positions[i]
		rect.X = pos.x
		rect.Y = pos.y
		if validate(layout.Rects, rect) {
			if len(layout.Rects) >= len(sandbox.Rects) {
				sandbox.Rects = append(sandbox.Rects, rect)
			} else {
				sandbox.Rects[len(layout.Rects)] = rect
			}

			var size = findBounds(sandbox.Rects)
			sandbox.Width = size.width
			sandbox.Height = size.height

			var score = rate(sandbox)
			if score < bestScore {
				bestScore = score
				bestRect.X = rect.X
				bestRect.Y = rect.Y
			}
		}
	}

	return bestRect
}

// determine order of iteration (FFD)
func preorder(sizes []Rect) []int {
	var order = make([]int, len(sizes))
	for i := 0; i < len(sizes); i++ {
		order[i] = i
	}

	// sort Rects by area descending

	sort.Slice(order, func(a, b int) bool {
		return sizes[b].Width*sizes[b].Height-sizes[a].Width*sizes[a].Height < 0
	})

	return order
}

// rearrange Rects to reflect the given iteration order
func reorder(items []Rect, order []int) []Rect {
	for i := 0; i < len(items); i++ {
		var tmp = items[order[i]]
		items[order[i]] = items[i]
		items[i] = tmp
	}
	return items
}

// packs { Width, Height } tuples into a layout { Width, Height, Rects }
func pack(sizes []Rect) Layout {
	var layout = Layout{
		Width:  0,
		Height: 0,
		Rects:  []Rect{},
	}

	if len(sizes) <= 0 {
		return layout
	}

	var order = preorder(sizes)
	for i := 0; i < len(sizes); i++ {
		var size = sizes[order[i]]

		var rect = findBestRect(layout, size)
		rect.Name = size.Name
		layout.Rects = append(layout.Rects, rect)

		var bounds = findBounds(layout.Rects)
		layout.Width = bounds.width
		layout.Height = bounds.height
	}

	reorder(layout.Rects, order)
	return layout
}
