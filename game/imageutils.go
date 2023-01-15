package game

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

func textureToImage(filename string) *image.RGBA {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	m := image.NewRGBA(p.Bounds())

	draw.Draw(m, m.Bounds(), p, image.ZP, draw.Src)

	return m
}
