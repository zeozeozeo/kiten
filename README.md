# kiten

Simple and fast 2D graphics library for Go.

# Examples

## Create a new canvas

```go
width := 1280
height := 720
canvas := kiten.NewCanvas(width, height, kiten.BlendAdd)
```

## Create a new canvas from an RGBA image

(you can use this later to overlay or draw on images, for example)

```go
canvas := kiten.CanvasFromImageRGBA(image, kiten.BlendAdd)
```

## [Real time example](https://github.com/zeozeozeo/kiten-simple)

![example](https://github.com/zeozeozeo/kiten-simple/blob/main/example.gif?raw=true)

## Fill the whole canvas with a black color

If you're rendering multiple frames, remember to clear the canvas when drawing a new frame

```go
canvas.Fill(color.RGBA{0, 0, 0, 255})
```

## Draw a red line from 100,100 to 500,700

```go
canvas.Line(100, 100, 500, 700, color.RGBA{255, 0, 0, 255})
```

## Draw a blue circle at 150,150 with the radius of 70

```go
canvas.Circle(150, 150, 70, color.RGBA{0, 0, 255, 255})
```

## Draw a filled blue circle at 150,150 with the radius of 70

```go
canvas.CircleFilled(150, 150, 70, color.RGBA{0, 0, 255, 255})
```

## Draw a filled blue circle with a green outline at 150,150 with the radius of 70

```go
canvas.CircleOutline(150, 150, 70, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255})
```

## Set the pixel at 500,500 to white

This will do nothing if the coordinates are invalid

```go
canvas.SetPixel(500, 500, color.RGBA{255, 255, 255, 255})
```

## Get the pixel value at 170,300

This will return a transparent color if the coordinates are invalid

```go
canvas.PixelAt(170, 300) // Returns color.RGBA
```

## Draw a green rectangle from 100,100 to 200,200 (not filled)

```go
canvas.Rect(100, 100, 200, 200, color.RGBA{0, 255, 0, 255})
```

## Draw a green rectangle from 100,100 to 200,200 (filled)

```go
canvas.RectFilled(100, 100, 200, 200, color.RGBA{0, 255, 0, 255})
```

## Draw a canvas on top of another canvas at 450,300 and resize it to 100 by 150 pixels

```go
canvas.PutCanvas(450, 300, 100, 150, canvas2)
```

## Connect a path with white lines

```go
import (
    "rand"
    "image"
)

// Generate a path
path := []image.Point{}
for i := 0; i < 100; i++ {
    path = append(path, image.Pt(i*30, canvas.Height/2-rand.Intn(100)))
}

// Draw it
canvas.DrawPath(path, color.RGBA{255, 255, 255, 255})
```

## Draw text at 500,500

```go
import "golang.org/x/image/font/inconsolata"
canvas.Text("Hello World!", 500, 500, inconsolata.Regular8x16, color.RGBA{255, 255, 255, 255})
```

## Export canvas to PNG

```go
import "os"
file, err := os.Create("output.png")
if err != nil {
    panic(err)
}
defer file.Close()

err := canvas.WritePNG(file)
if err != nil {
    panic(err)
}
```

## Draw a white (not filled) triangle (500,500 -> 500,700 -> 600,700)

```go
canvas.Triangle(500, 500, 500, 700, 600, 700, color.RGBA{255, 255, 255, 255})
```

## Draw a white filled triangle (500,500 -> 500,700 -> 600,700)

```go
canvas.TriangleFilled(500, 500, 500, 700, 600, 700, color.RGBA{255, 255, 255, 255})
```

## Draw a white filled triangle with a red outline (500,500 -> 500,700 -> 600,700)

```go
canvas.TriangleOutline(500, 500, 500, 700, 600, 700, color.RGBA{255, 255, 255, 255}, color.RGBA{255, 0, 0, 255})
```
