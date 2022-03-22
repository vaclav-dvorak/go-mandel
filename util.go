package main

import (
	"image/color"
	"math"

	"github.com/vaclav-dvorak/go-mandel/palette"
)

func calcColor(val float64) color.RGBA {
	pal := palette.ColorPalettes[conf.Palette]
	if val > float64(conf.Iterations) {
		return pal[len(pal)-1]
	}
	i, frac := math.Modf((float64(len(pal)-1) * val) / float64(conf.Iterations))
	return cosineInterpolation(pal[int(i)], pal[int(i)+1], frac)
}

func cosineInterpolation(c1, c2 color.RGBA, mu float64) color.RGBA {
	sr, sg, sb, sa := c1.RGBA()
	tr, tg, tb, _ := c2.RGBA()
	mu2 := (1 - math.Cos(mu*math.Pi)) / 2.0
	rr := uint8(float64(sr)*(1-mu2) + float64(tr)*mu2)
	rg := uint8(float64(sg)*(1-mu2) + float64(tg)*mu2)
	rb := uint8(float64(sb)*(1-mu2) + float64(tb)*mu2)
	return color.RGBA{rr, rg, rb, uint8(sa)}
}
