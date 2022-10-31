package kiten

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Canvas struct {
	Image     *image.RGBA
	Pixels    int
	Width     int
	Height    int
	BlendType BlendType
}

type BlendType int

const (
	BlendAdd      BlendType = 0 // Add pixel values
	BlendMultiply BlendType = 1 // Multiply pixel values
	BlendNone     BlendType = 2 // Overwrite pixel values
)

// Create a new canvas with width of x and height of y
func NewCanvas(x, y int, blendType BlendType) *Canvas {
	canvas := &Canvas{BlendType: blendType}
	canvas.Image = image.NewRGBA(image.Rect(0, 0, x, y))
	canvas.Pixels = x * y
	canvas.Width, canvas.Height = x, y

	return canvas
}

// Convert an image into a canvas
func CanvasFromImageRGBA(img *image.RGBA, blendType BlendType) *Canvas {
	return &Canvas{
		Image:     img,
		Pixels:    img.Rect.Dx() * img.Rect.Dy(),
		Width:     img.Rect.Dx(),
		Height:    img.Rect.Dy(),
		BlendType: blendType,
	}
}

// Set a pixel on a canvas
func (canvas *Canvas) SetPixel(x, y int, color color.RGBA) {
	// Faster than image.Set
	pixelStart := (y-canvas.Image.Rect.Min.Y)*canvas.Image.Stride + (x-canvas.Image.Rect.Min.X)*4
	if pixelStart+3 > (canvas.Pixels*4)-1 || pixelStart < 0 {
		return
	}

	canvas.Image.Pix[pixelStart+3] = 255
	if color.A == 255 || canvas.BlendType == BlendNone {
		// Set pixel value
		canvas.Image.Pix[pixelStart] = color.R
		canvas.Image.Pix[pixelStart+1] = color.G
		canvas.Image.Pix[pixelStart+2] = color.B
	} else if canvas.BlendType == BlendMultiply {
		alphaFloat := float32(color.A)
		canvas.Image.Pix[pixelStart] += uint8(float32(color.R) * (alphaFloat / 255))
		canvas.Image.Pix[pixelStart+1] += uint8(float32(color.G) * (alphaFloat / 255))
		canvas.Image.Pix[pixelStart+2] += uint8(float32(color.B) * (alphaFloat / 255))
	} else if canvas.BlendType == BlendAdd {
		canvas.Image.Pix[pixelStart] += color.R
		canvas.Image.Pix[pixelStart+1] += color.G
		canvas.Image.Pix[pixelStart+2] += color.B
	}
}

// Returns the color value at x and y
func (canvas *Canvas) PixelAt(x, y int) color.RGBA {
	pixelStart := (y-canvas.Image.Rect.Min.Y)*canvas.Image.Stride + (x-canvas.Image.Rect.Min.X)*4
	if pixelStart+3 > (canvas.Pixels*4) || pixelStart < 0 {
		return color.RGBA{}
	}

	return color.RGBA{
		canvas.Image.Pix[pixelStart],
		canvas.Image.Pix[pixelStart+1],
		canvas.Image.Pix[pixelStart+2],
		canvas.Image.Pix[pixelStart+3],
	}
}

// Fills the canvas with a color
func (canvas *Canvas) Fill(color color.RGBA) {
	for x := 0; x < canvas.Width; x++ {
		for y := 0; y < canvas.Height; y++ {
			canvas.SetPixel(x, y, color)
		}
	}
}

// Draws a line
func (canvas *Canvas) Line(x1, y1, x2, y2 int, color color.RGBA) {
	// TODO: Antialiasing
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}
	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy
	x, y := x1, y1

	for {
		canvas.SetPixel(x, y, color)
		if x == x2 && y == y2 {
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

// Draws a filled rectangle
func (canvas *Canvas) RectFilled(x1, y1, x2, y2 int, color color.RGBA) {
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			canvas.SetPixel(x, y, color)
		}
	}
}

// Draws a rectangle (not filled)
func (canvas *Canvas) Rect(x1, y1, x2, y2 int, color color.RGBA) {
	canvas.Line(x1, y1, x2, y1, color) // Left to right, top
	canvas.Line(x2, y1, x2, y2, color) // Top to bottom, right
	canvas.Line(x1, y2, x2, y2, color) // Left to right, bottom
	canvas.Line(x1, y1, x1, y2, color) // Top to bottom, left
}

// Draws a circle (not filled)
func (canvas *Canvas) Circle(cx, cy, r int, color color.RGBA) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x >= y {
		// TODO: Clip circle when x < 0
		if cx+x >= canvas.Width || cx-x < 0 {
			return
		}
		canvas.SetPixel(cx+x, cy+y, color)
		canvas.SetPixel(cx+y, cy+x, color)
		canvas.SetPixel(cx-y, cy+x, color)
		canvas.SetPixel(cx-x, cy+y, color)
		canvas.SetPixel(cx-x, cy-y, color)
		canvas.SetPixel(cx-y, cy-x, color)
		canvas.SetPixel(cx+y, cy-x, color)
		canvas.SetPixel(cx+x, cy-y, color)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}

// Draws a filled circle
func (canvas *Canvas) CircleFilled(cx, cy, r int, color color.RGBA) {
	floatR := float64(r)

	for x := -r; x <= r; x++ {
		floatX := float64(x)
		height := int(math.Sqrt(floatR*floatR - floatX*floatX))

		// Get rid of the extra pixel at the top and bottom
		if height != r {
			for y := -height; y < height; y++ {
				canvas.SetPixel(x+cx, y+cy, color)
			}
		} else {
			for y := -height + 1; y < height-1; y++ {
				canvas.SetPixel(x+cx, y+cy, color)
			}
		}
	}
}

// Draws a circle with an outline
func (canvas *Canvas) CircleOutline(cx, cy, r int, insideColor, outlineColor color.RGBA) {
	canvas.CircleFilled(cx, cy, r, insideColor)
	canvas.Circle(cx, cy, r, outlineColor)
}

// Draw a canvas on top of this canvas
func (canvas *Canvas) PutCanvas(x, y, w, h int, canvas2 *Canvas) {
	if canvas2.Width == 0 || canvas2.Height == 0 || canvas.Width == 0 || canvas.Height == 0 || w == 0 || h == 0 {
		return
	}

	// Scaling factor
	scaleX := float64(canvas2.Width) / float64(w)
	scaleY := float64(canvas2.Height) / float64(h)

	for ox := 0; ox < w; ox++ {
		for oy := 0; oy < h; oy++ {
			sx := int(float64(ox) * scaleX)
			sy := int(float64(oy) * scaleY)
			canvas.SetPixel(ox+x, oy+y, canvas2.PixelAt(sx, sy))
		}
	}
}

// Connects a path of points with lines
func (canvas *Canvas) DrawPath(path []image.Point, color color.RGBA) {
	for i := 0; i < len(path)-1; i++ {
		point := path[i]
		nextPoint := path[i+1]
		canvas.Line(point.X, point.Y, nextPoint.X, nextPoint.Y, color)
	}
}

func (canvas *Canvas) Text(text string, x, y int, face *basicfont.Face, color color.RGBA) {
	textCanvas := NewCanvas(len(text)*face.Width, face.Height, BlendAdd)
	point := fixed.Point26_6{X: fixed.I(0), Y: fixed.I(face.Height)}
	drawer := &font.Drawer{
		Dst:  textCanvas.Image,
		Src:  image.NewUniform(color),
		Face: face,
		Dot:  point,
	}

	drawer.DrawString(text)
	// TODO: Text resizing
	canvas.PutCanvas(
		x,
		y,
		textCanvas.Width,
		textCanvas.Height,
		textCanvas,
	)
}

// Convert radians to degrees
func Rad2Deg(radians float64) float64 {
	return radians * (180 / math.Pi)
}

// Convert degrees to radians
func Deg2Rad(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

// Rotate a point in space by degrees
func (canvas *Canvas) RotatePoint(x, y, degrees float64) (int, int) {
	halfWidth := float64(canvas.Width) / 2
	halfHeight := float64(canvas.Height) / 2

	dx := x - halfWidth
	dy := y - halfHeight
	mag := math.Sqrt(dx*dx + dy*dy)
	dir := math.Atan2(dy, dx) + Deg2Rad(degrees)
	return int(math.Cos(dir)*mag + halfWidth), int(math.Sin(dir)*mag + halfHeight)
}

// Draws a triangle (not filled)
func (canvas *Canvas) Triangle(x1, y1, x2, y2, x3, y3 int, color color.RGBA) {
	canvas.Line(x1, y1, x2, y2, color)
	canvas.Line(x2, y2, x3, y3, color)
	canvas.Line(x3, y3, x1, y1, color)
}

// Draws a filled triangle
func (canvas *Canvas) TriangleFilled(x1, y1, x2, y2, x3, y3 int, color color.RGBA) {
	// Swap
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if y2 > y3 {
		x2, x3 = x3, x2
		y2, y3 = y3, y2
	}
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	dx12 := x2 - x1
	dy12 := y2 - y1
	dx13 := x3 - x1
	dy13 := y3 - y1

	for y := y1; y <= y2; y++ {
		if 0 <= y && y < canvas.Height {
			var s1, s2 int
			if dy12 != 0 {
				s1 = (y-y1)*dx12/dy12 + x1
			} else {
				s1 = x1
			}
			if dy13 != 0 {
				s2 = (y-y1)*dx13/dy13 + x1
			} else {
				s2 = x1
			}
			if s1 > s2 {
				s1, s2 = s2, s1
			}

			for x := s1; x <= s2; x++ {
				if 0 <= x && x < canvas.Width {
					canvas.SetPixel(x, y, color)
				}
			}
		}
	}

	dx32 := x2 - x3
	dy32 := y2 - y3
	dx31 := x1 - x3
	dy31 := y1 - y3

	for y := y2; y <= y3; y++ {
		if 0 <= y && y < canvas.Height {
			var s1, s2 int
			if dy32 != 0 {
				s1 = (y-y3)*dx32/dy32 + x3
			} else {
				s1 = x3
			}
			if dy31 != 0 {
				s2 = (y-y3)*dx31/dy31 + x3
			} else {
				s2 = x3
			}
			if s1 > s2 {
				s1, s2 = s2, s1
			}

			for x := s1; x <= s2; x++ {
				if 0 <= x && x < canvas.Width {
					canvas.SetPixel(x, y, color)
				}
			}
		}
	}
}

// Draws a triangle with an outline
func (canvas *Canvas) TriangleOutline(x1, y1, x2, y2, x3, y3 int, fillColor, outlineColor color.RGBA) {
	canvas.TriangleFilled(x1, y1, x2, y2, x3, y3, fillColor)
	canvas.Triangle(x1, y1, x2, y2, x3, y3, outlineColor)
}

// Is the point inside of the canvas?
func (canvas *Canvas) IsPointInCanvas(x, y int) bool {
	return x >= 0 && x < canvas.Width && y >= 0 && y < canvas.Height
}

// Export to PNG
func (canvas *Canvas) WritePNG(writer io.Writer) error {
	return png.Encode(writer, canvas.Image)
}
