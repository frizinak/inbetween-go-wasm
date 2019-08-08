package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"sync"
	"syscall/js"
	"time"

	"github.com/frizinak/inbetween-go-wasm/carve"
	"golang.org/x/image/bmp"
)

func imgString(img *image.RGBA) string {
	buf := bytes.NewBuffer([]byte("data:image/bmp;base64,"))
	w := base64.NewEncoder(base64.StdEncoding, buf)
	if err := bmp.Encode(w, img); err != nil {
		panic(err)
	}
	w.Close()
	return buf.String()
}

// func render(window, body, canvas, ctx js.Value, img *image.RGBA) {
func render(canvas js.Value, img *image.RGBA) <-chan struct{} {
	// imgEl := window.Get("Image").New()
	// imgEl.Set("style", "display: none")
	// imgEl.Set("src", imgString(img))

	wait := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// canvas.Set("width", imgEl.Get("width"))
		// canvas.Set("height", imgEl.Get("height"))
		// ctx.Call("drawImage", imgEl, 0, 0)
		// body.Call("removeChild", imgEl)
		wait <- struct{}{}
		cb.Release()
		return nil
	})

	canvas.Set("onload", cb)
	canvas.Set("src", imgString(img))
	// imgEl.Set("onload", cb)
	// body.Call("appendChild", imgEl)
	return wait
}

func main() {
	img, err := carve.Img("castle.png")
	if err != nil {
		panic(err)
	}

	window := js.Global()
	document := window.Get("document")
	canvas := document.Call("getElementById", "canvas")

	//carveIV := time.Millisecond
	var rw sync.Mutex

	done := false

	var anim js.Func
	anim = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			rw.Lock()
			// render(window, body, canvas, ctx, img)
			wait := render(canvas, img)
			rw.Unlock()
			<-wait
			if done {
				anim.Release()
				return
			}
			window.Call("requestAnimationFrame", anim)
		}()
		return nil
	})

	go func() {
		window.Call("requestAnimationFrame", anim)
	}()

	go func() {
		for {
			rw.Lock()

			carve.Vert(img)
			if img.Bounds().Dy() <= 280 {
				done = true
				rw.Unlock()
				break
			}

			rw.Unlock()
			time.Sleep(time.Microsecond * 50)
		}

	}()

	c := make(chan struct{})
	<-c
}
