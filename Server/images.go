package main

import (
	"bytes"
	"encoding/base64"
	"image"
	colour "image/color"
	"image/png"
	"strings"
)

const (
	// The number of pixels wide for the border between blocks
	borderWidth = 1
)

// GenImage returns the base64 encoding of the image if success, error if there was an error
func GenImage(req imageReq) (string, error) {
	img := image.NewRGBA(
		image.Rectangle{
			image.Point{
				0,
				0,
			},
			image.Point{
				req.Width*req.BlockWidth + req.Width*borderWidth,
				len(req.BoardPositions)*req.BlockHeight + len(req.BoardPositions)*borderWidth,
			},
		},
	)

	borderColourRGBA := colour.RGBA{uint8(req.BorderColour[0]), uint8(req.BorderColour[1]), uint8(req.BorderColour[2]), 255}

	for rowNum, row := range req.BoardPositions {
		for cellNum, cell := range row {
			cell = strings.ToLower(cell)

			var RGBColour [3]int

			switch cell {
			case " ":
				// Blank Space
				RGBColour = req.BackGroundColour
			case "a":
				// Apple space
				RGBColour = req.AppleColour
			case "h":
				// Head of the snake
				RGBColour = req.HeadColour
			case "s":
				// Body of snake
				RGBColour = req.BodyColour
			default:
				// Unknown, use background colour
				RGBColour = req.BackGroundColour
			}

			RGBAColour := colour.RGBA{uint8(RGBColour[0]), uint8(RGBColour[1]), uint8(RGBColour[2]), 255}

			minX := (cellNum * req.BlockWidth) + (cellNum * borderWidth)

			maxX := ((cellNum + 1) * req.BlockWidth) + (cellNum * borderWidth)

			minY := (rowNum * req.BlockHeight) + (rowNum * borderWidth)

			maxY := ((rowNum + 1) * req.BlockHeight) + (rowNum * borderWidth)

			for y := maxY; y > minY; y-- {
				for x := minX; x < maxX; x++ {
					img.Set(x, y, RGBAColour)
				}
				img.Set(maxX+1, y, borderColourRGBA)
			}
		}
	}

	// Fill any pixels with an alpha value of 0 (indicating not set) with the border colour
	for x := img.Bounds().Min.X; x <= img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y <= img.Bounds().Max.Y; y++ {
			if _, _, _, a := img.At(x, y).RGBA(); a == 0 {
				img.Set(x, y, borderColourRGBA)
			}
		}
	}

	buffer := new(bytes.Buffer)

	err := png.Encode(buffer, img)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}
