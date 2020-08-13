package main

import (
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: gopher-pic [src image] [gopher image] [output image]")
	}

	src := os.Args[1]
	gopher := os.Args[2]
	out := os.Args[3]

	srcImage := gocv.IMRead(src, gocv.IMReadUnchanged)
	if srcImage.Empty() {
		log.Fatal("invalid source image")
	}

	gopherImage := gocv.IMRead(gopher, gocv.IMReadUnchanged)
	if gopherImage.Empty() {
		log.Fatal("invalid gopher image")
	}

	log.Println("init detector")
	hog := gocv.NewHOGDescriptor()
	if err := hog.SetSVMDetector(gocv.HOGDefaultPeopleDetector()); err != nil {
		log.Fatal("detector init failed")
	}

	log.Println("resize source image")
	maxSize := math.Max(float64(srcImage.Cols()), float64(srcImage.Rows()))
	scale := 1.0
	srcResized := srcImage
	if maxSize > 500 {
		scale = 1.0 / (maxSize / 500.0)
		srcResized = gocv.NewMat()
		gocv.Resize(srcImage, &srcResized, image.Point{}, scale, scale, gocv.InterpolationLinear)
	}

	log.Println("scale: ", scale, " resized size: ", srcResized.Rows(), "x", srcResized.Cols())

	log.Println("detecting human")
	rects := hog.DetectMultiScaleWithParams(
		srcResized,
		0,
		image.Point{0, 0},
		image.Point{0, 0},
		1.05,
		2.0,
		false,
	)
	if len(rects) == 0 {
		log.Fatal("no human on source image")
	}
	humanRect := rects[0]
	for _, rect := range rects {
		rect.Union(humanRect)
	}
	if !(math.Abs(scale-1.0) < 1e-9) {
		humanRect = image.Rect(
			int(float64(humanRect.Min.X)*(1/scale)),
			int(float64(humanRect.Min.Y)*(1/scale)),
			int(float64(humanRect.Max.X)*(1/scale)),
			int(float64(humanRect.Max.Y)*(1/scale)),
		)
	}
	log.Println("human rect found", humanRect)

	ratio := float64(humanRect.Dy()) / float64(gopherImage.Rows())
	needSize := float64(humanRect.Dy()) / 2.0
	gopherScale := 1.0
	if !(math.Abs(ratio-1) < 1e-9) {
		gopherScale = needSize / float64(gopherImage.Rows())
	}
	gocv.Resize(gopherImage, &gopherImage, image.Point{}, gopherScale, gopherScale, gocv.InterpolationCubic)

	log.Println("gopher scale and new size", gopherScale, gopherImage.Size())

	layers := gocv.Split(gopherImage)
	rgb := layers[0:3]
	mask := layers[3]
	gocv.Merge(rgb, &gopherImage)

	dst := srcImage.RowRange(
		humanRect.Max.Y-gopherImage.Rows(),
		humanRect.Max.Y,
	)
	dst = dst.ColRange(
		humanRect.Max.X-humanRect.Dx()/2,
		humanRect.Max.X-humanRect.Dx()/2+gopherImage.Cols(),
	)
	gopherImage.CopyToWithMask(&dst, mask)

	gocv.IMWrite(out, srcImage)
}
