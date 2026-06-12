//go:build ignore

// gen-icons generates PWA icons (192x192 and 512x512) into web/icons/
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

func main() {
	os.MkdirAll("web/icons", 0755)
	genIcon(192, "web/icons/icon-192.png")
	genIcon(512, "web/icons/icon-512.png")
}

func genIcon(size int, path string) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{88, 86, 214, 255} // indigo #5856d6

	// fill background with rounded feel (square PNG, maskable)
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// draw a simple "S" letter in white
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size) * 0.28
	white := color.RGBA{255, 255, 255, 255}

	// draw circle outline as a shield shape
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist >= r-float64(size)*0.04 && dist <= r {
				img.Set(x, y, white)
			}
		}
	}

	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}
