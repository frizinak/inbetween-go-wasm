package main

import (
	"os"

	"github.com/frizinak/inbetween-go-wasm/carve"
	"golang.org/x/image/bmp"
)

func main() {
	img, err := carve.Img("castle.png")
	if err != nil {
		panic(err)
	}

	for {
		carve.Vert(img)
		if img.Bounds().Dy() <= 280 {
			break
		}
	}

	out, err := os.Create("carve-result.bmp")
	if err != nil {
		panic(err)
	}
	if err := bmp.Encode(out, img); err != nil {
		panic(err)
	}
}
