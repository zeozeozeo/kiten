package kiten

import (
	"image/color"
)

// Draws a line
func (canvas *Canvas) Line(x0 int, y0 int, x1 int, y1 int, color color.RGBA) {
	// TODO: Antialiasing (Xiaolin Wu's line algorithm)
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy
	x, y := x0, y0

	for {
		canvas.SetPixel(x, y, color)
		if x == x1 && y == y1 {
			return
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
			if x >= canvas.Width || x < 0 {
				return
			}
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}
