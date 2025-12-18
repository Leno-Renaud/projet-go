package main

import "testing"

func fillBlock(img [][]Pixel, bx, by, factor int, p Pixel) {
	for y := by; y < by+factor; y++ {
		for x := bx; x < bx+factor; x++ {
			img[y][x] = p
		}
	}
}

func TestTransformToTarget_SimpleMapping(t *testing.T) {
	width, height, factor := 4, 4, 2

	// Create source image with four distinct uniform blocks
	src := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		src[y] = make([]Pixel, width)
	}

	red := Pixel{R: 0xffff, G: 0x0, B: 0x0}
	green := Pixel{R: 0x0, G: 0xffff, B: 0x0}
	blue := Pixel{R: 0x0, G: 0x0, B: 0xffff}
	gray := Pixel{R: 30000, G: 30000, B: 30000}

	// source layout:
	// (0,0)=red  | (2,0)=green
	// (0,2)=blue | (2,2)=gray
	fillBlock(src, 0, 0, factor, red)
	fillBlock(src, 2, 0, factor, green)
	fillBlock(src, 0, 2, factor, blue)
	fillBlock(src, 2, 2, factor, gray)

	// Create target image where blocks are permutations of the source blocks
	tgt := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		tgt[y] = make([]Pixel, width)
	}

	// target layout:
	// (0,0)=green  | (2,0)=red
	// (0,2)=gray   | (2,2)=blue
	fillBlock(tgt, 0, 0, factor, green)
	fillBlock(tgt, 2, 0, factor, red)
	fillBlock(tgt, 0, 2, factor, gray)
	fillBlock(tgt, 2, 2, factor, blue)

	result := transformToTarget(src, tgt, width, height, factor)

	// Verify each target block in result contains the pixels from the expected source block
	expectedAt := map[[2]int]Pixel{
		{0, 0}: green, // target (0,0) should be mapped to source green
		{2, 0}: red,
		{0, 2}: gray,
		{2, 2}: blue,
	}

	for by := 0; by < height; by += factor {
		for bx := 0; bx < width; bx += factor {
			exp := expectedAt[[2]int{bx, by}]
			for y := by; y < by+factor; y++ {
				for x := bx; x < bx+factor; x++ {
					if result[y][x] != exp {
						t.Fatalf("unexpected pixel at (%d,%d): got %+v want %+v for block (%d,%d)", x, y, result[y][x], exp, bx, by)
					}
				}
			}
		}
	}
}

func TestTransformToTarget_NonDivisibleDimensions(t *testing.T) {
	width, height, factor := 5, 5, 3

	// Create source with a 3x3 red block at (0,0)
	src := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		src[y] = make([]Pixel, width)
	}
	red := Pixel{R: 0xffff, G: 0x0, B: 0x0}
	fillBlock(src, 0, 0, factor, red)

	// Create target with a small block at bottom-right (2x2)
	tgt := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		tgt[y] = make([]Pixel, width)
	}
	// place a target block at (3,3) with color equal to red so matching chooses the 3x3 red
	fillBlock(tgt, 3, 3, 2, red)

	// Should not panic; result must have same dimensions
	result := transformToTarget(src, tgt, width, height, factor)
	if len(result) != height || len(result[0]) != width {
		t.Fatalf("unexpected result size: got %dx%d want %dx%d", len(result[0]), len(result), width, height)
	}

	// The bottom-right 2x2 area (3..4,3..4) should be filled with red (clipped from the 3x3 source)
	for y := 3; y < height; y++ {
		for x := 3; x < width; x++ {
			if result[y][x] != red {
				t.Fatalf("expected red at (%d,%d), got %+v", x, y, result[y][x])
			}
		}
	}
}
