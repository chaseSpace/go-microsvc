package ufile

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/h2non/filetype/matchers"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

func CompressImageFile(buf []byte) ([]byte, bool, error) {
	if is, ext, err := IsImageFile(buf, matchers.TypeJpeg, matchers.TypePng); err != nil {
		return nil, false, errors.New(fmt.Sprintf("IsImageFile failed: %v", err))
	} else if !is {
		return nil, false, errors.New(fmt.Sprintf("not image(jpeg/png) file: %s", ext))
	} else {
		imgSrc, _, err := image.Decode(bytes.NewReader(buf)) // image: unknown format
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Decode image failed: %v", err))
		}
		newImg := image.NewRGBA(imgSrc.Bounds())
		if ext == "png" { // reformat to jpeg
			draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
			draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
		}
		buf2 := bytes.Buffer{}
		err = jpeg.Encode(&buf2, newImg, &jpeg.Options{Quality: 40})
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Encode image failed: %v", err))
		}
		if buf2.Len() > len(buf) {
			return buf, false, nil
		}
		return buf2.Bytes(), true, nil
	}
}
