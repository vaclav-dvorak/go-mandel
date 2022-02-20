package main

import (
	"image/color"
	"math"

	"github.com/vaclav-dvorak/go-mandelbrot/palette"
)

func calcColor(val float64) color.RGBA {
	pal := palette.ColorPalettes[conf.Palette]
	if val > float64(conf.Iterations) {
		return pal[len(pal)-1]
	}
	i, frac := math.Modf((float64(len(pal)-1) * val) / float64(conf.Iterations))
	sr, sg, sb, sa := pal[int(i)].RGBA()
	tr, tg, tb, _ := pal[int(i)+1].RGBA()
	return color.RGBA{cosineInterpolation(float64(sr), float64(tr), frac), cosineInterpolation(float64(sg), float64(tg), frac), cosineInterpolation(float64(sb), float64(tb), frac), uint8(sa)}
}

func cosineInterpolation(c1, c2, mu float64) uint8 {
	mu2 := (1 - math.Cos(mu*math.Pi)) / 2.0
	return uint8(c1*(1-mu2) + c2*mu2)
}
