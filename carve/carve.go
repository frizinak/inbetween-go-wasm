package carve

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"runtime"

	"github.com/frizinak/inbetween-go-wasm/bound"
)

func Img(asset string) (*image.RGBA, error) {
	d, err := bound.Asset(asset)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(d)
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return img.(*image.RGBA), nil
}

func colorDist(r, g, b, r1, g1, b1 uint8) float64 {
	rv := float64(r) - float64(r1)
	gv := float64(g) - float64(g1)
	bv := float64(b) - float64(b1)
	return math.Sqrt(rv*rv + gv*gv + bv*bv)
}

func colorDistPix(pix []uint8, ix0, ix1 int) float64 {
	if ix1 <= 0 || ix1 >= len(pix) {
		return math.MaxFloat64
	}

	return colorDist(
		pix[ix0+0],
		pix[ix0+1],
		pix[ix0+2],
		pix[ix1+0],
		pix[ix1+1],
		pix[ix1+2],
	)
}

func find(img *image.RGBA) (float64, []int) {
	bounds := img.Bounds()
	dx := bounds.Dx()

	var x, y, ry, ix0, ix1, ix2, ix3 int
	var dist, dist1, dist2, dist3 float64
	var lastTotal = math.MaxFloat64
	var total float64

	cline := make([]int, dx)
	line := make([]int, dx)

	for y = bounds.Min.Y; y < bounds.Max.Y; y++ {
		ry = y
		ix0 = img.PixOffset(bounds.Min.X, y)
		cline[0] = ix0
		total = 0.0
		// amount = 0
		for x = bounds.Min.X; x < bounds.Max.X-1; x++ {
			ix0 = img.PixOffset(x, ry)
			ix1 = img.PixOffset(x+1, ry)
			ix2 = img.PixOffset(x+1, ry+1)
			ix3 = img.PixOffset(x+1, ry-1)
			dist1 = colorDistPix(img.Pix, ix0, ix1)
			dist2 = colorDistPix(img.Pix, ix0, ix2)
			dist3 = colorDistPix(img.Pix, ix0, ix3)

			ix0 = ix1
			dist = dist1
			if dist2 < dist1 && dist2 < dist3 {
				ix0 = ix2
				dist = dist2
				ry = ry + 1
			} else if dist3 < dist1 && dist3 < dist2 {
				ix0 = ix3
				dist = dist3
				ry = ry - 1
			}

			// total = ((total * amount) + dist) / (amount + 1)
			// amount++
			// total += dist
			if dist > total {
				total = dist
				if total > lastTotal {
					break
				}
			}

			ix0 += bounds.Min.Y * img.Stride
			cline[x+1] = ix0
		}

		if total < lastTotal {
			lastTotal = total
			copy(line, cline)
		}
	}

	return lastTotal, line
}

type result struct {
	score float64
	line  []int
}

func Vert(img *image.RGBA) {
	bounds := img.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	workers := runtime.NumCPU()
	results := make(chan result, workers)
	n := dy / workers
	for i := 0; i < workers; i++ {
		go func(img *image.RGBA) {
			score, line := find(img)
			results <- result{score, line}
		}(img.SubImage(image.Rect(0, i*n, dx, i*n+n)).(*image.RGBA))
	}

	var score = math.MaxFloat64
	var line []int
	for i := 0; i < workers; i++ {
		result := <-results
		if result.score < score {
			score = result.score
			line = result.line
		}
	}
	close(results)

	for _, v := range line {
		for i := v; i < dy*img.Stride-img.Stride; i += img.Stride {
			img.Pix[i+0] = img.Pix[i+img.Stride+0]
			img.Pix[i+1] = img.Pix[i+img.Stride+1]
			img.Pix[i+2] = img.Pix[i+img.Stride+2]
		}
	}

	img.Rect.Max.Y--
}
