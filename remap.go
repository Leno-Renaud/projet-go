package main

import (
	"math/rand"
)

// Remap pixels from a source image to match the color distribution of a target image
// without changing pixel values—only their positions. Images must share dimensions.

// quantizePixel maps a Pixel to a bin index using `levels` discrete values per channel.
// Example: levels=16 → 4096 bins. Pixels are not modified; this is only for grouping.
func quantizePixel(p Pixel, levels int) int {
	step := 65536 / levels
	qR := int(p.R) / step
	qG := int(p.G) / step
	qB := int(p.B) / step
	if qR >= levels {
		qR = levels - 1
	}
	if qG >= levels {
		qG = levels - 1
	}
	if qB >= levels {
		qB = levels - 1
	}
	return (qR*levels+qG)*levels + qB
}

// binCenter returns the approximate center value (0..65535) of a bin on each channel.
func binCenter(binIdx, levels int) (int, int, int) {
	step := 65536 / levels
	plane := levels * levels
	qR := binIdx / plane
	rem := binIdx % plane
	qG := rem / levels
	qB := rem % levels
	cR := qR*step + step/2
	cG := qG*step + step/2
	cB := qB*step + step/2
	return cR, cG, cB
}

func sqDist(aR, aG, aB, bR, bG, bB int) int {
	dR := aR - bR
	dG := aG - bG
	dB := aB - bB
	return dR*dR + dG*dG + dB*dB
}

// buildSourceBins groups source pixels into bins and keeps their exact values.
func buildSourceBins(src [][]Pixel, levels int) [][]Pixel {
	bins := make([][]Pixel, levels*levels*levels)
	for y := 0; y < len(src); y++ {
		row := src[y]
		for x := 0; x < len(row); x++ {
			p := row[x]
			bin := quantizePixel(p, levels)
			bins[bin] = append(bins[bin], p)
		}
	}
	return bins
}

// sortPixelsByDistance sorts a slice of pixels by squared distance to (cR, cG, cB),
// closest first. Uses a simple insertion sort for small slices.
func sortPixelsByDistance(pixels []Pixel, cR, cG, cB int) {
	// Simple insertion sort (fine for typical bin sizes).
	for i := 1; i < len(pixels); i++ {
		key := pixels[i]
		keyDist := sqDist(int(key.R), int(key.G), int(key.B), cR, cG, cB)
		j := i - 1
		for j >= 0 && sqDist(int(pixels[j].R), int(pixels[j].G), int(pixels[j].B), cR, cG, cB) > keyDist {
			pixels[j+1] = pixels[j]
			j--
		}
		pixels[j+1] = key
	}
}

// buildTargetHistogram counts how many pixels of the target fall into each bin.
func buildTargetHistogram(target [][]Pixel, levels int) []int {
	hist := make([]int, levels*levels*levels)
	for y := 0; y < len(target); y++ {
		row := target[y]
		for x := 0; x < len(row); x++ {
			bin := quantizePixel(row[x], levels)
			hist[bin]++
		}
	}
	return hist
}

// popPixel removes and returns one pixel from the requested bin; if empty, it finds
// the nearest bin (in quantized color space) that still has supply. Returns false if
// no pixel is available (should not happen when images have identical pixel counts).
func popPixel(bin int, bins [][]Pixel, levels int) (Pixel, bool) {
	if len(bins[bin]) > 0 {
		last := bins[bin][len(bins[bin])-1]
		bins[bin] = bins[bin][:len(bins[bin])-1]
		return last, true
	}

	// Fallback: search nearest bin with remaining supply.
	targetCR, targetCG, targetCB := binCenter(bin, levels)
	bestIdx := -1
	bestDist := int(^uint(0) >> 1) // max int

	for idx, supply := range bins {
		if len(supply) == 0 {
			continue
		}
		cR, cG, cB := binCenter(idx, levels)
		dist := sqDist(targetCR, targetCG, targetCB, cR, cG, cB)
		if dist < bestDist {
			bestDist = dist
			bestIdx = idx
		}
	}

	if bestIdx == -1 {
		return Pixel{}, false
	}

	last := bins[bestIdx][len(bins[bestIdx])-1]
	bins[bestIdx] = bins[bestIdx][:len(bins[bestIdx])-1]
	return last, true
}

// remapPixels rearranges source pixels to match the target color distribution.
// Assumptions: src and target have identical dimensions. No pixel value is changed.
// levels controls the number of bins per channel (e.g., 16 → 4096 bins).
// Pixels are placed in a randomized order to distribute source pixels uniformly.
func remapPixels(src [][]Pixel, target [][]Pixel, levels int) [][]Pixel {
	if len(src) == 0 || len(target) == 0 || len(src) != len(target) || len(src[0]) != len(target[0]) {
		return nil
	}

	height := len(target)
	width := len(target[0])

	// Supply: group all source pixels by bin.
	bins := buildSourceBins(src, levels)

	// Output image (same dimensions).
	out := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		out[y] = make([]Pixel, width)
	}

	// Create a randomized order of target positions to avoid bias.
	positions := make([][2]int, height*width)
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			positions[idx] = [2]int{x, y}
			idx++
		}
	}
	// Shuffle positions.
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Fill the output image in randomized order.
	for _, pos := range positions {
		x, y := pos[0], pos[1]
		bin := quantizePixel(target[y][x], levels)
		p, ok := popPixel(bin, bins, levels)
		if !ok {
			// Should not happen; keep original target pixel as a safety.
			out[y][x] = target[y][x]
			continue
		}
		out[y][x] = p
	}

	return out
}

