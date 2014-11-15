package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
)

const (
	PatternCanvas = "pattern"
	MotifCanvas   = "motif"
)

func openImg(filename string, verbose bool) image.Image {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func readStitches(img image.Image, canvas string, verbose bool) string {
	width  := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	if verbose {
		fmt.Fprintf(os.Stderr, "  Width : %4d\n  Height: %4d\n", width, height)
	}

	if canvas == PatternCanvas && (width != 20 || height != 270) {
		log.Fatalf("Dimensions for canvas %s must be 20 x 270, got %d x %d", canvas, width, height)
	} else if canvas == MotifCanvas && (width != 80 || height != 270) {
		log.Fatalf("Dimensions for canvas %s must be 80 x 270, got %d x %d", canvas, width, height)
	}

	pixels := make([]string, 0, width*height)

	blackStitches := 0
	whiteStitches := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			isAlpha := (a >> 8) < 128
			isBlack := ((r>>8 + g>>8 + b>>8) / 3) < 128

			// Skip transparent and white in pattern
			if !(isAlpha || (canvas == PatternCanvas && !isBlack)) {
				pixels = append(pixels, fmt.Sprintf("\"%d:%d\":%t", x+1, y+1, isBlack))

				if isBlack {
					blackStitches++
				} else {
					whiteStitches++
				}
			}
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Black stitches: %5d\n  White stitches: %5d\n", blackStitches, whiteStitches)
	}

	return strings.Join(pixels, ",")
}

func parseImg(filename string, canvas string, verbose bool) string {
	if canvas != PatternCanvas && canvas != MotifCanvas {
		log.Fatalf("Unknown canvas \"%s\"", canvas)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Reading %s \"%s\"\n", canvas, filename)
	}

	img := openImg(filename, verbose)
	stitches := readStitches(img, canvas, verbose)

	return fmt.Sprintf("knit.pixels.%s = JSON.parse('{%s}'); knit.drawPixels(knit.%sContext, knit.pixels.%s);",
		canvas,stitches, canvas, canvas)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-v] pattern.png motif.png\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
	}

	pattern := parseImg(flag.Args()[0], PatternCanvas, *verbose)
	motif   := parseImg(flag.Args()[1], MotifCanvas,   *verbose)

	fmt.Fprintf(os.Stdout, "knit.resetCanvas(); %s; %s; knit.repeatPattern(); knit.preview(false, true);", pattern, motif)
}
