package main

import (
	"image"
	"runtime"
	"sync"
)

// extractPixelsParallel convertit une image en matrice de pixels RGB (version parallélisée)
func extractPixelsParallel(m image.Image, width, height int) [][]Pixel {
	rgbMatrix := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		rgbMatrix[y] = make([]Pixel, width)
	}

	numWorkers := runtime.NumCPU()
	chunkSize := height / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		startY := i * chunkSize
		endY := startY + chunkSize
		if i == numWorkers-1 {
			endY = height
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := 0; x < width; x++ {
					r, g, b, _ := m.At(x, y).RGBA()
					rgbMatrix[y][x] = Pixel{R: uint16(r), G: uint16(g), B: uint16(b)}
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return rgbMatrix
}

// blackWhiteParallel convertit la matrice en niveaux de gris (parallèle, in-place)
func blackWhiteParallel(rgbMatrix [][]Pixel, width, height int) [][]Pixel {
	numGoroutines := runtime.NumCPU()
	rowsPerGoroutine := (height + numGoroutines - 1) / numGoroutines

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()

			startRow := goroutineID * rowsPerGoroutine
			endRow := (goroutineID + 1) * rowsPerGoroutine
			if endRow > height {
				endRow = height
			}

			for y := startRow; y < endRow; y++ {
				for x := 0; x < width; x++ {
					p := rgbMatrix[y][x]
					gray := uint16(0.299*float64(p.R) + 0.587*float64(p.G) + 0.114*float64(p.B))
					rgbMatrix[y][x] = Pixel{R: gray, G: gray, B: gray}
				}
			}
		}(g)
	}

	wg.Wait()
	return rgbMatrix
}
