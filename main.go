package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"

	"github.com/vaclav-dvorak/go-mandelbrot/palette"
)

type config struct {
	Real       map[string]float64 `koanf:"real"`
	Imag       map[string]float64 `koanf:"imag"`
	Palette    string             `koanf:"palette"`
	Radius     float64            `koanf:"radius"`
	Width      int                `koanf:"width"`
	Height     int                `koanf:"height"`
	Step       int                `koanf:"step"`
	Iterations int                `koanf:"iteration"`
	Workers    int                `koanf:"workers"`
}

type pix struct {
	x int
	y int
	c color.RGBA
}

var (
	conf         = config{}
	k            = koanf.New(".")
	version      = "dev"
	steps        map[string]int
	img          *image.RGBA
	foundDivider int
)

func run() {
	img = image.NewRGBA(image.Rect(0, 0, conf.Width, conf.Height))
	cfg := pixelgl.WindowConfig{
		Title:       fmt.Sprintf("Mandelbrot set z^2+c @%s", version),
		Bounds:      pixel.R(0, 0, float64(conf.Width), float64(conf.Height)),
		VSync:       true,
		Undecorated: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatalf("error creating window: %v", err)
	}

	drawBuffer := make(chan pix, (conf.Width/steps["x"])*(conf.Height/steps["y"]))
	go drawThread(drawBuffer)
	go workersInit(drawBuffer)
	for !win.Closed() {
		if win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ) {
			return
		}

		pic := pixel.PictureDataFromImage(img)
		sprite := pixel.NewSprite(pic, pic.Bounds())
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	if err := k.Unmarshal("mandel", &conf); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	if _, ok := palette.ColorPalettes[conf.Palette]; !ok {
		log.Fatalf("pallet %s not found", conf.Palette)
	}
	pal := palette.ColorPalettes[conf.Palette]
	log.Printf("%v\n", pal)
	steps = getSteps()
	pixelgl.Run(run)
}

func getSteps() map[string]int {
	dividers := []int{}
	for i := 3; i < (conf.Width/2)+1; i++ {
		if conf.Width%i == 0 && conf.Height%i == 0 {
			dividers = append(dividers, i)
		}
	}
	mid := dividers[len(dividers)/2 : len(dividers)/2+1][0] //? mid is slice of one element. so we pick it up directly with [0]
	foundDivider = mid
	return map[string]int{"x": conf.Width / mid, "y": conf.Height / mid}
}

func drawThread(pixs <-chan pix) {
	for i := range pixs {
		img.SetRGBA(i.x, i.y, i.c)
	}
}

func workersInit(pixs chan<- pix) {
	var wg sync.WaitGroup
	start := time.Now()
	workerBuffer := make(chan map[string]int, conf.Workers)
	for w := 0; w < conf.Workers; w++ {
		wg.Add(1)
		go func(wb <-chan map[string]int) {
			realw := math.Abs(conf.Real["from"] - conf.Real["to"])
			imagw := math.Abs(conf.Imag["from"] - conf.Imag["to"])
			for work := range wb {
				for x := work["fromx"]; x < work["tox"]; x++ {
					for y := work["fromy"]; y < work["toy"]; y++ {
						rx := realw*float64(x)/float64(conf.Width) + conf.Real["from"]
						ry := imagw*float64(y)/float64(conf.Height) + conf.Imag["from"]
						ch, i := mandelbrotIteraction(rx, ry)
						if ch > 4 {
							pixs <- pix{x: x, y: y, c: hslToRGB(float64(i)/100*ch, 1, 0.5)}
						} else {
							pixs <- pix{x: x, y: y, c: color.RGBA{R: 0, G: 0, B: 0, A: 255}}
						}

					}
				}
			}
			wg.Done()
		}(workerBuffer)
	}
	for x := 0; x <= conf.Width; x += steps["x"] {
		for y := 0; y <= conf.Height; y += steps["y"] {
			workerBuffer <- map[string]int{"fromx": x, "tox": x + steps["x"], "fromy": y, "toy": y + steps["y"]}
		}
	}
	close(workerBuffer)
	wg.Wait()
	log.Printf("\nworkers: %d\ndivider: %d\ntook: %s", conf.Workers, foundDivider, time.Since(start))
	close(pixs)
}

func mandelbrotIteraction(a, b float64) (float64, int) {
	var x, y, xx, yy, xy float64

	for i := 0; i < conf.Iterations; i++ {
		xx, yy, xy = x*x, y*y, x*y
		if xx+yy > 4 {
			return xx + yy, i
		}
		// xn+1 = x^2 - y^2 + a
		x = xx - yy + a
		// yn+1 = 2xy + b
		y = 2*xy + b
	}

	return xx + yy, conf.Iterations
}
