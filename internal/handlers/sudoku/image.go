package sudoku

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// Grid constants (all in pixels).
const (
	cellSize  = 64
	padding   = 10
	thinLine  = 1
	thickLine = 3
	digitW    = 5  // bitmap columns
	digitH    = 7  // bitmap rows
	digitPx   = 7  // pixels per bitmap dot → 35×49 digit in 64px cell
)

// cellPos returns the top-left pixel coordinate of row/column n.
func cellPos(n int) int {
	p := padding + thickLine
	for i := 0; i < n; i++ {
		p += cellSize
		if (i+1)%3 == 0 {
			p += thickLine
		} else {
			p += thinLine
		}
	}
	return p
}

func imgSize() int {
	return cellPos(8) + cellSize + thickLine + padding
}

// 5×7 pixel bitmaps for digits 1-9 (index 0 is unused/blank).
// Each uint8 represents one row; bit 4 = leftmost column.
var digitBitmaps = [10][7]uint8{
	0: {},
	1: {0b00100, 0b01100, 0b00100, 0b00100, 0b00100, 0b00100, 0b01110},
	2: {0b01110, 0b10001, 0b00001, 0b00110, 0b01000, 0b10000, 0b11111},
	3: {0b11110, 0b00001, 0b00001, 0b01110, 0b00001, 0b00001, 0b11110},
	4: {0b00010, 0b00110, 0b01010, 0b10010, 0b11111, 0b00010, 0b00010},
	5: {0b11111, 0b10000, 0b10000, 0b11110, 0b00001, 0b00001, 0b11110},
	6: {0b01110, 0b10000, 0b10000, 0b11110, 0b10001, 0b10001, 0b01110},
	7: {0b11111, 0b00001, 0b00010, 0b00100, 0b01000, 0b01000, 0b01000},
	8: {0b01110, 0b10001, 0b10001, 0b01110, 0b10001, 0b10001, 0b01110},
	9: {0b01110, 0b10001, 0b10001, 0b01111, 0b00001, 0b00001, 0b01110},
}

var (
	colBlack  = color.RGBA{0, 0, 0, 255}
	colWhite  = color.RGBA{255, 255, 255, 255}
	colGray   = color.RGBA{180, 180, 180, 255}
	colBlue   = color.RGBA{30, 100, 200, 255}
	colCellBG = color.RGBA{245, 248, 255, 255} // very light blue for solved cells
)

// RenderPuzzle draws the 81-char puzzle string (0=empty).
// For the solution image, pass puzzle and solution; given cells are black, solved are blue.
func RenderPuzzle(puzzle, solution string) ([]byte, error) {
	sz := imgSize()
	img := image.NewRGBA(image.Rect(0, 0, sz, sz))
	draw.Draw(img, img.Bounds(), &image.Uniform{colWhite}, image.Point{}, draw.Src)

	// Draw digits
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			idx := r*9 + c
			puzz := puzzle[idx] - '0'
			sol := byte(0)
			if solution != "" {
				sol = solution[idx] - '0'
			}

			cx := cellPos(c)
			cy := cellPos(r)

			if puzz != 0 {
				// Given cell: black number, white bg
				drawDigit(img, cx, cy, int(puzz), colBlack)
			} else if sol != 0 {
				// Solved cell: blue number, light blue bg
				fillRect(img, cx, cy, cellSize, cellSize, colCellBG)
				drawDigit(img, cx, cy, int(sol), colBlue)
			}
		}
	}

	drawGridLines(img, sz)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawDigit(img *image.RGBA, cx, cy, d int, clr color.Color) {
	if d < 1 || d > 9 {
		return
	}
	bmp := digitBitmaps[d]
	startX := cx + (cellSize-digitW*digitPx)/2
	startY := cy + (cellSize-digitH*digitPx)/2

	for row := 0; row < digitH; row++ {
		for col := 0; col < digitW; col++ {
			if (bmp[row]>>uint(digitW-1-col))&1 == 1 {
				for py := 0; py < digitPx; py++ {
					for px := 0; px < digitPx; px++ {
						img.Set(startX+col*digitPx+px, startY+row*digitPx+py, clr)
					}
				}
			}
		}
	}
}

func drawGridLines(img *image.RGBA, sz int) {
	lineLen := sz - 2*padding

	// Vertical lines (before each column + right edge)
	for c := 0; c <= 9; c++ {
		var x, w int
		var clr color.Color
		if c == 0 {
			x, w, clr = padding, thickLine, colBlack
		} else if c == 9 {
			x, w, clr = cellPos(8)+cellSize, thickLine, colBlack
		} else if c%3 == 0 {
			x, w, clr = cellPos(c)-thickLine, thickLine, colBlack
		} else {
			x, w, clr = cellPos(c)-thinLine, thinLine, colGray
		}
		fillRect(img, x, padding, w, lineLen, clr)
	}

	// Horizontal lines (before each row + bottom edge)
	for r := 0; r <= 9; r++ {
		var y, h int
		var clr color.Color
		if r == 0 {
			y, h, clr = padding, thickLine, colBlack
		} else if r == 9 {
			y, h, clr = cellPos(8)+cellSize, thickLine, colBlack
		} else if r%3 == 0 {
			y, h, clr = cellPos(r)-thickLine, thickLine, colBlack
		} else {
			y, h, clr = cellPos(r)-thinLine, thinLine, colGray
		}
		fillRect(img, padding, y, lineLen, h, clr)
	}
}

func fillRect(img *image.RGBA, x, y, w, h int, clr color.Color) {
	r := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, r, &image.Uniform{clr}, image.Point{}, draw.Src)
}
