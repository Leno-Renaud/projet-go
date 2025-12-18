package main

import (
	"testing"
)

// helper pour copier une matrice sans partager la mémoire
func copyMatrix(src [][]Pixel) [][]Pixel {
	h := len(src)
	if h == 0 {
		return nil
	}
	w := len(src[0])
	dst := make([][]Pixel, h)
	for y := 0; y < h; y++ {
		dst[y] = make([]Pixel, w)
		copy(dst[y], src[y])
	}
	return dst
}

func BenchmarkExtractPixels(b *testing.B) {
	m := loadImage("image.jpg")
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPixels(m, width, height)
	}
}

func BenchmarkExtractPixelsParallel(b *testing.B) {
	m := loadImage("image.jpg")
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPixelsParallel(m, width, height)
	}
}

func BenchmarkBlackWhiteSeq(b *testing.B) {
	m := loadImage("image.jpg")
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	base := extractPixels(m, width, height)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mat := copyMatrix(base)
		_ = blackWhiteSeq(mat, width, height)
	}
}

func BenchmarkBlackWhiteParallel(b *testing.B) {
	m := loadImage("image.jpg")
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	base := extractPixels(m, width, height)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mat := copyMatrix(base)
		_ = blackWhiteParallel(mat, width, height)
	}
}
