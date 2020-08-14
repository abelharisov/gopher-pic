package main

import (
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"os"
)

func main() {
	// получаем аргументы программы
	src, gopher, out := args()
	// загружаем исходное изображение и картинку с гофером
	srcImage, gopherImage := loadImages(src, gopher)
	// определяем, где на картинке находится человек
	humanRect := detectHuman(srcImage)
	// масштабируем гофера, что бы его высота была примерно 1/5 от человека
	gopherScaledImage := scaleGopher(humanRect, gopherImage)
	// рисуем гофера на картинке
	resultImage := drawGopher(gopherScaledImage, srcImage, humanRect)

	// пишем результат в файл
	gocv.IMWrite(out, resultImage)
}

func drawGopher(gopherImage gocv.Mat, srcImage gocv.Mat, humanRect image.Rectangle) gocv.Mat {
	// нужно выделить слой с альфа каналом и использовать его как маску при копировании изображения гофера
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

	return srcImage
}

func scaleGopher(humanRect image.Rectangle, gopherImage gocv.Mat) gocv.Mat {
	ratio := float64(humanRect.Dy()) / float64(gopherImage.Rows())
	needSize := float64(humanRect.Dy()) / 2.0
	gopherScale := 1.0
	if !(math.Abs(ratio-1) < 1e-9) {
		gopherScale = needSize / float64(gopherImage.Rows())
	}
	gocv.Resize(gopherImage, &gopherImage, image.Point{}, gopherScale, gopherScale, gocv.InterpolationCubic)
	return gopherImage
}

func detectHuman(srcImage gocv.Mat) (humanRect image.Rectangle) {
	hog := gocv.NewHOGDescriptor()
	if err := hog.SetSVMDetector(gocv.HOGDefaultPeopleDetector()); err != nil {
		log.Fatal("detector init failed")
	}

	// уменьшаем исходное изображение, что бы ускорить работу алгоритма поиска человека
	maxSize := math.Max(float64(srcImage.Cols()), float64(srcImage.Rows()))
	scale := 1.0
	srcResized := srcImage
	if maxSize > 500 {
		scale = 1.0 / (maxSize / 500.0)
		srcResized = gocv.NewMat()
		gocv.Resize(srcImage, &srcResized, image.Point{}, scale, scale, gocv.InterpolationLinear)
	}

	rects := hog.DetectMultiScale(srcResized)
	if len(rects) == 0 {
		log.Fatal("no human on source image")
	}

	// если нашлось несколько областей, то объединяем в одну
	humanRect = rects[0]
	for _, rect := range rects {
		rect.Union(humanRect)
	}

	// если исходное  изображение масштабировалось, то нужно отмасштабироть найденную область обратно
	if !(math.Abs(scale-1.0) < 1e-9) {
		humanRect = image.Rect(
			int(float64(humanRect.Min.X)*(1/scale)),
			int(float64(humanRect.Min.Y)*(1/scale)),
			int(float64(humanRect.Max.X)*(1/scale)),
			int(float64(humanRect.Max.Y)*(1/scale)),
		)
	}
	return humanRect
}

func loadImages(src string, gopher string) (srcImage, gopherImage gocv.Mat) {
	srcImage = gocv.IMRead(src, gocv.IMReadUnchanged)
	if srcImage.Empty() {
		log.Fatal("invalid source image")
	}
	gopherImage = gocv.IMRead(gopher, gocv.IMReadUnchanged)
	if gopherImage.Empty() {
		log.Fatal("invalid gopher image")
	}
	return
}

func args() (src, gopher, out string) {
	if len(os.Args) < 4 {
		log.Fatal("Usage: gopher-pic [src image] [gopher image] [output image]")
	}
	src = os.Args[1]
	gopher = os.Args[2]
	out = os.Args[3]
	return
}
