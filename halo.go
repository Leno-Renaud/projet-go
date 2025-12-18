package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Indispensable pour décoder le JPEG (init function)
	"image/png"
	_ "image/png" // Indispensable pour décoder le PNG (init function)
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// Pixel représente une valeur RGB
type Pixel struct {
	R, G, B uint16
}

// loadImage charge une image depuis un fichier
func loadImage(filename string) image.Image {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

// extractPixels convertit une image en matrice de pixels RGB
func extractPixels(m image.Image, width, height int) [][]Pixel {
	rgbMatrix := make([][]Pixel, height)

	for y := 0; y < height; y++ {
		rgbMatrix[y] = make([]Pixel, width)

		for x := 0; x < width; x++ {
			r, g, b, _ := m.At(x, y).RGBA()

			rgbMatrix[y][x] = Pixel{
				R: uint16(r),
				G: uint16(g),
				B: uint16(b),
			}
		}
	}
	return rgbMatrix
}

// extractPixelsParallel convertit une image en matrice de pixels RGB (version parallélisée)
func extractPixelsParallel(m image.Image, width, height int) [][]Pixel {
	// Initialiser la matrice avec allocation sécurisée
	rgbMatrix := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		rgbMatrix[y] = make([]Pixel, width)
	}

	// Déterminer le nombre de workers
	numWorkers := runtime.NumCPU()
	chunkSize := height / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	// WaitGroup pour synchronisation
	var wg sync.WaitGroup

	// Créer et lancer les workers
	for i := 0; i < numWorkers; i++ {
		startY := i * chunkSize
		endY := startY + chunkSize
		if i == numWorkers-1 {
			endY = height
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			// Boucle sur les lignes assignées à ce worker
			for y := start; y < end; y++ {
				// Boucle sur tous les pixels de la ligne
				for x := 0; x < width; x++ {
					// Lire le pixel depuis l'image source (thread-safe)
					r, g, b, _ := m.At(x, y).RGBA()

					// Écrire dans sa portion de matrice (pas de race condition)
					rgbMatrix[y][x] = Pixel{
						R: uint16(r),
						G: uint16(g),
						B: uint16(b),
					}
				}
			}
		}(startY, endY)
	}

	// Attendre que tous les workers terminent
	wg.Wait()

	return rgbMatrix
}

func blackWhite(rgbMatrix [][]Pixel, width, height int) [][]Pixel {
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

// pixelsToImage convertit une matrice de pixels en image RGBA
func pixelsToImage(rgbMatrix [][]Pixel, width, height int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := rgbMatrix[y][x]

			out.Set(x, y, color.RGBA{
				R: uint8(p.R >> 8),
				G: uint8(p.G >> 8),
				B: uint8(p.B >> 8),
				A: 255,
			})
		}
	}

	return out
}

// saveImage sauvegarde une image en PNG
func saveImage(img *image.RGBA, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Charger l'image
	m := loadImage("image.jpg")
	// Récupérer les dimensions
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	fmt.Printf("Dimensions : %dx%d\n", width, height)
	fmt.Printf("Nombre de cœurs : %d\n\n", runtime.NumCPU())

	// ============================================
	// Test 1 : extractPixels (VERSION SÉQUENTIELLE)
	// ============================================
	fmt.Println("=== TEST extractPixels (SÉQUENTIEL) ===")
	start1 := time.Now()
	rgbMatrix1 := extractPixels(m, width, height)
	duration1 := time.Since(start1)
	fmt.Printf("Temps : %v\n\n", duration1)

	// ============================================
	// Test 2 : extractPixelsParallel (VERSION PARALLÈLE)
	// ============================================
	fmt.Println("=== TEST extractPixelsParallel (PARALLÈLE) ===")
	start2 := time.Now()
	rgbMatrix2 := extractPixelsParallel(m, width, height)
	duration2 := time.Since(start2)
	fmt.Printf("Temps : %v\n\n", duration2)

	// ============================================
	// Comparaison
	// ============================================
	fmt.Println("=== COMPARAISON ===")
	speedup := float64(duration1) / float64(duration2)
	savings := (1 - float64(duration2)/float64(duration1)) * 100
	fmt.Printf("Speedup : %.2fx\n", speedup)
	fmt.Printf("Gain de temps : %.2f%%\n", savings)
	fmt.Printf("Différence : %v\n\n", duration1-duration2)

	// Utiliser la version parallèle pour le résultat final
	rgbMatrix := rgbMatrix2
	_ = rgbMatrix1 // éviter l'avertissement "unused"
	rgbMatrix = blackWhite(rgbMatrix, width, height)
	// Convertir en image RGBA
	out := pixelsToImage(rgbMatrix, width, height)
	// Sauvegarder
	saveImage(out, "out.png")
	fmt.Println("Image sauvegardée : out.png")
}
