package main

import (
	"gocv.io/x/gocv"
	"log"
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

	log.Println("detector inited")

	log.Println("detecting human")

	rects := hog.DetectMultiScale(srcImage)
	if len(rects) == 0 {
		log.Fatal("no human on source image")
	}
	union := rects[0]
	for _, rect := range rects {
		rect.Union(union)
	}

	log.Println("human rect found", union)

	// todo
	// заскейлить исходное изображение до 1000 px с любой стороны
	// заскейлить гофера, что бы он занимал 1/5 высоты области с человеком
	// разместить гофера на уровне ног

	layers := gocv.Split(gopherImage)
	rgb := layers[0:3]
	mask := layers[3]
	gocv.Merge(rgb, &gopherImage)

	dst := srcImage.RowRange(0, gopherImage.Rows())
	dst = dst.ColRange(union.Max.X, union.Max.X+gopherImage.Cols())
	gopherImage.CopyToWithMask(&dst, mask)

	gocv.IMWrite(out, srcImage)

	//window := gocv.NewWindow("Hello")
	//window.IMShow(dst)
	//window.WaitKey(1)
}
