package main

import (
	"bytes"
	"encoding/base64"
	"image"
	colour "image/color" // Spelt correctly now
	"image/png"
	_ "image/png"
	"strings"
)

const (
	BORDER_WIDTH = 1
)

func GenImage(req ImageReq) (string, error) {
	img := image.NewRGBA(
		image.Rectangle{
			image.Point{
				0,
				0,
			},
			image.Point{
				req.width*req.BlockWidth + req.width*BORDER_WIDTH,
				len(req.BoardPositions)*req.BlockHeight + len(req.BoardPositions)*BORDER_WIDTH,
			},
		},
	)

	border_colour_rgba := colour.RGBA{uint8(req.BorderColour[0]), uint8(req.BorderColour[1]), uint8(req.BorderColour[2]), 255}

	for row_num, row := range req.BoardPositions {
		for cell_num, cell := range row {
			cell = strings.ToLower(cell)

			var rgb_colour [3]int

			if cell == " " {
				// Blank Space
				rgb_colour = req.BackGroundColour
			} else if cell == "a" {
				// Apple space
				rgb_colour = req.AppleColour
			} else if cell == "h" {
				// Head of snake
				rgb_colour = req.HeadColour
			} else if cell == "s" {
				// Body of snake
				rgb_colour = req.BodyColour
			} else {
				// Unknown, use background colour
				rgb_colour = req.BackGroundColour
			}

			rgba_colour := colour.RGBA{uint8(rgb_colour[0]), uint8(rgb_colour[1]), uint8(rgb_colour[2]), 255}

			min_x := (cell_num * req.BlockWidth) + (cell_num * BORDER_WIDTH)

			max_x := ((cell_num + 1) * req.BlockWidth) + (cell_num * BORDER_WIDTH)

			min_y := (row_num * req.BlockHeight) + (row_num * BORDER_WIDTH)

			max_y := ((row_num + 1) * req.BlockHeight) + (row_num * BORDER_WIDTH)

			for y := max_y; y > min_y; y-- {
				for x := min_x; x < max_x; x++ {
					img.Set(x, y, rgba_colour)
				}
				img.Set(max_x+1, y, border_colour_rgba)
			}
		}
	}

	for x := img.Bounds().Min.X; x <= img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y <= img.Bounds().Max.Y; y++ {
			if _, _, _, a := img.At(x, y).RGBA(); a == 0 {
				img.Set(x, y, border_colour_rgba)
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
