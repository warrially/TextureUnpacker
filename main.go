package main

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/image/bmp"

	"github.com/nfnt/resize"
	"github.com/yangtizi/go/ioutils"
	"github.com/yangtizi/log/color"
)

type TFile struct {
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	W       int    `json:"w"`
	H       int    `json:"h"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	OffX    int    `json:"offX"`
	OffY    int    `json:"offY"`
	SourceW int    `json:"sourceW"`
	SourceH int    `json:"sourceH"`
}

type TJson struct {
	File       string           `json:"file"`
	ImagePath  string           `json:"imagePath"`
	Frame      map[string]TFile `json:"frames"`
	SubTexture []TFile          `json:"SubTexture"`
}

func main() {
	fmt.Println("version 2021年9月1日 11:43:23")
	files, err := ioutil.ReadDir(`.`)
	if err != nil {
		panic(err)
	}
	// 获取文件，并输出它们的名字
	for _, file := range files {
		if ioutils.Path.GetExtension(file.Name()) == ".json" {
			println("正在解析文件", file.Name())
			main1(file.Name())
		}
	}
}

func egret(v *TJson) {
	src := v.File
	strExt := ioutils.Path.GetExtension(src)
	ioutils.Directory.CreateDirectory("output/")

	for k2, v2 := range v.Frame {

		dst := "output/" + k2 + strExt
		fIn, _ := os.Open(src)
		defer fIn.Close()

		fOut, _ := os.Create(dst)
		defer fOut.Close()

		fmt.Printf("正在切图 %s%s", k2, color.Red)
		err := Clip(fIn, fOut, 0, 0, v2.X, v2.Y, v2.X+v2.W, v2.Y+v2.H, 100)
		if err != nil {
			panic(err)
		}

		fmt.Println(color.Green + "      [OK]" + color.Reset)

	}

}

func dragonbone(v *TJson) {
	src := v.ImagePath
	strExt := ioutils.Path.GetExtension(src)
	ioutils.Directory.CreateDirectory("output/")

	for _, v2 := range v.SubTexture {
		fmt.Println(v2.X, v2.Y, v2.Width, v2.Height)
		dst := "output/" + v2.Name + strExt
		fIn, _ := os.Open(src)
		defer fIn.Close()

		fOut, _ := os.Create(dst)
		defer fOut.Close()

		fmt.Printf("正在切图 %s%s", v2.Name, color.Red)
		err := Clip(fIn, fOut, 0, 0, v2.X, v2.Y, v2.X+v2.Width, v2.Y+v2.Height, 100)
		if err != nil {
			panic(err)
		}

		fmt.Println(color.Green + "      [OK]" + color.Reset)

	}

}

func main1(strJSON string) {

	// ioutils.Path.

	v := &TJson{}

	ioutils.File.JSON(strJSON, v)

	// fmt.Println(v)

	if ioutils.File.Exists(v.File) {
		egret(v)
		return
	}

	if ioutils.File.Exists(v.ImagePath) {
		fmt.Println("龙骨切", v.ImagePath)
		dragonbone(v)
	}

	/*	w, h, err := Scale(fIn, fOut, 150, 150, 100)
		if err != nil {
			panic(err)
		}
		fmt.Println(w, h)*/
}

//补上缺失的代码
//* Clip 图片裁剪
//* 入参:图片输入、输出、缩略图宽、缩略图高、Rectangle{Pt(x0, y0), Pt(x1, y1)}，精度
//* 规则:如果精度为0则精度保持不变
//*
//* 返回:error
// */
func Clip(in io.Reader, out io.Writer, wi, hi, x0, y0, x1, y1, quality int) (err error) {
	err = errors.New("unknow error")
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	var origin image.Image
	var fm string
	origin, fm, err = image.Decode(in)
	if err != nil {
		log.Println(err)
		return err
	}

	if wi == 0 || hi == 0 {
		wi = origin.Bounds().Max.X
		hi = origin.Bounds().Max.Y
	}
	var canvas image.Image
	if wi != origin.Bounds().Max.X {
		//先缩略
		canvas = resize.Thumbnail(uint(wi), uint(hi), origin, resize.Lanczos3)
	} else {
		canvas = origin
	}

	switch fm {
	case "jpeg":
		img := canvas.(*image.YCbCr)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)
		return jpeg.Encode(out, subImg, &jpeg.Options{quality})
	case "png":
		switch canvas.(type) {
		case *image.NRGBA:
			img := canvas.(*image.NRGBA)
			rect := image.Rect(x0, y0, x1, y1)
			subImg := img.SubImage(rect).(*image.NRGBA)
			subImg.PixOffset(100, 100)
			return png.Encode(out, subImg)
		case *image.RGBA:
			img := canvas.(*image.RGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
			return png.Encode(out, subImg)
		}
	case "gif":
		img := canvas.(*image.Paletted)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.Paletted)
		return gif.Encode(out, subImg, &gif.Options{})
	case "bmp":
		img := canvas.(*image.RGBA)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
		return bmp.Encode(out, subImg)
	default:
		return errors.New("ERROR FORMAT")
	}
	return nil
}

/*
* Scale 缩略图生成
* 入参:图片输入、输出，缩略图宽、高，精度
* 规则: 如果width 或 hight其中有一个为0，则大小不变 如果精度为0则精度保持不变
* 返回:缩略图真实宽、高、error

 */
func Scale(in io.Reader, out io.Writer, width, height, quality int) (int, int, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	var (
		w, h int
	)
	origin, fm, err := image.Decode(in)
	if err != nil {
		log.Println(err)
		return 0, 0, err
	}
	if width == 0 || height == 0 {
		width = origin.Bounds().Max.X
		height = origin.Bounds().Max.Y
	}
	if quality == 0 {
		quality = 100
	}
	canvas := resize.Thumbnail(uint(width), uint(height), origin, resize.Lanczos3)

	//return jpeg.Encode(out, canvas, &jpeg.Options{quality})
	w = canvas.Bounds().Dx()
	h = canvas.Bounds().Dy()
	switch fm {
	case "jpeg":
		return w, h, jpeg.Encode(out, canvas, &jpeg.Options{quality})
	case "png":
		return w, h, png.Encode(out, canvas)
	case "gif":
		return w, h, gif.Encode(out, canvas, &gif.Options{})
	//case "bmp":  //被我注释掉的是x/image/bmp
	//	return w, h, bmp.Encode(out, canvas)
	default:
		return w, h, errors.New("ERROR FORMAT")
	}
	return w, h, nil
}
