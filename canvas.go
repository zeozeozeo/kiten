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
func NewCanvas(x int, y int, blendType BlendType) *Canvas {
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
func (canvas *Canvas) SetPixel(x int, y int, color color.RGBA) {
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
func (canvas *Canvas) PixelAt(x int, y int) color.RGBA {
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
func (canvas *Canvas) Line(x0 int, y0 int, x1 int, y1 int, color color.RGBA) {
	// TODO: Antialiasing
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

// Draws a filled rectangle
func (canvas *Canvas) RectFilled(x1 int, y1 int, x2 int, y2 int, color color.RGBA) {
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			canvas.SetPixel(x, y, color)
		}
	}
}

// Draws a rectangle (not filled)
func (canvas *Canvas) Rect(x1 int, y1 int, x2 int, y2 int, color color.RGBA) {
	canvas.Line(x1, y1, x2, y1, color) // Left to right, top
	canvas.Line(x2, y1, x2, y2, color) // Top to bottom, right
	canvas.Line(x1, y2, x2, y2, color) // Left to right, bottom
	canvas.Line(x1, y1, x1, y2, color) // Top to bottom, left
}

// Draws a circle (not filled)
func (canvas *Canvas) Circle(cx int, cy int, r int, color color.RGBA) {
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

// Draws a filled cirlce
func (canvas *Canvas) CircleFilled(cx int, cy int, r int, color color.RGBA) {
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
func (canvas *Canvas) CircleOutline(cx int, cy int, r int, insideColor color.RGBA, outlineColor color.RGBA) {
	canvas.CircleFilled(cx, cy, r, insideColor)
	canvas.Circle(cx, cy, r, outlineColor)
}

// Draw a canvas on top of this canvas
func (canvas *Canvas) PutCanvas(x int, y int, w int, h int, canvas2 *Canvas) {
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

func (canvas *Canvas) Text(text string, x int, y int, face *basicfont.Face, color color.RGBA) {
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
func (canvas *Canvas) RotatePoint(x float64, y float64, degrees float64) (int, int) {
	halfWidth := float64(canvas.Width) / 2
	halfHeight := float64(canvas.Height) / 2

	dx := x - halfWidth
	dy := y - halfHeight
	mag := math.Sqrt(dx*dx + dy*dy)
	dir := math.Atan2(dy, dx) + Deg2Rad(degrees)
	return int(math.Cos(dir)*mag + halfWidth), int(math.Sin(dir)*mag + halfHeight)
}

// Is the point inside of the canvas?
func (canvas *Canvas) IsPointInCanvas(x int, y int) bool {
	return x > 0 && x < canvas.Width && y > 0 && y < canvas.Height
}

// Export to PNG
func (canvas *Canvas) WritePNG(writer io.Writer) error {
	return png.Encode(writer, canvas.Image)
}
