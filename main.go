package main

import (
	"fmt"
	"runtime"
	"time"
)

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
	rgbMatrix = blackWhiteParallel(rgbMatrix, width, height)
	// Re-extraction (conserver le flux exact de l'ancien main)
	rgbMatrix = extractPixels(m, width, height)
	// Pixelisation
	rgbMatrix = downscalePixels(rgbMatrix, width, height, 4)

	// Convertir en image RGBA & sauvegarder
	out := pixelsToImage(rgbMatrix)
	saveImage(out, "out.png")
	fmt.Println("Image sauvegardée : out.png")
}
