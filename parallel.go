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

// downscalePixels réduit la définition sans changer la taille (pixelisation)
func downscalePixelsParallel(rgbMatrix [][]Pixel, width, height, factor int) [][]Pixel {
	if factor <= 1 {
		return rgbMatrix
	}
	if len(rgbMatrix) == 0 || len(rgbMatrix[0]) == 0 {
		return rgbMatrix
	}

	result := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		result[y] = make([]Pixel, width)
	}

	// Diviser le travail par lignes
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

			// Traiter les blocs dans cette tranche de lignes
			for by := startRow; by < endRow; by += factor {
				for bx := 0; bx < width; bx += factor {
					var sumR, sumG, sumB uint64
					count := 0
					maxY := by + factor
					if maxY > height {
						maxY = height
					}
					maxX := bx + factor
					if maxX > width {
						maxX = width
					}

					// Calculer la moyenne du bloc
					for y := by; y < maxY; y++ {
						for x := bx; x < maxX; x++ {
							p := rgbMatrix[y][x]
							sumR += uint64(p.R)
							sumG += uint64(p.G)
							sumB += uint64(p.B)
							count++
						}
					}

					avg := Pixel{
						R: uint16(sumR / uint64(count)),
						G: uint16(sumG / uint64(count)),
						B: uint16(sumB / uint64(count)),
					}

					// Remplir le bloc avec la moyenne
					for y := by; y < maxY; y++ {
						for x := bx; x < maxX; x++ {
							result[y][x] = avg
						}
					}
				}
			}
		}(g)
	}

	wg.Wait()
	return result
}