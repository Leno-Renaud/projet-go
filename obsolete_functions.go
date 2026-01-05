package main

// Fonctions héritées (obsolètes) de remap par blocs spatiaux.
// Elles ne sont plus utilisées dans `main.go`, remplacées par `remapPixels` (remap par bins couleur).
// Conservées uniquement pour référence et pour les tests unitaires existants.

// Block représente un bloc `factor x factor` extrait d'une image.
// X, Y sont les coordonnées du coin supérieur gauche du bloc.
// Pixels contient les lignes du bloc, Avg est la couleur moyenne du bloc.
type Block struct {
	X, Y   int
	Pixels [][]Pixel
	Avg    Pixel
}

// colorDistance calcule une distance euclidienne au carré entre deux pixels (RGB 16 bits).
func colorDistance(a, b Pixel) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return dr*dr + dg*dg + db*db
}

// averageBlock retourne la couleur moyenne d'un bloc de pixels.
func averageBlock(pixels [][]Pixel) Pixel {
	var r, g, b uint64
	count := 0

	for _, row := range pixels {
		for _, p := range row {
			r += uint64(p.R)
			g += uint64(p.G)
			b += uint64(p.B)
			count++
		}
	}

	return Pixel{
		R: uint16(r / uint64(count)),
		G: uint16(g / uint64(count)),
		B: uint16(b / uint64(count)),
	}
}

// splitIntoBlocks découpe l'image en blocs factor×factor (le dernier bloc peut être plus petit).
func splitIntoBlocks(img [][]Pixel, width, height, factor int) []Block {
	var blocks []Block

	for by := 0; by < height; by += factor {
		for bx := 0; bx < width; bx += factor {

			maxY := by + factor
			if maxY > height {
				maxY = height
			}
			maxX := bx + factor
			if maxX > width {
				maxX = width
			}

			blockPixels := make([][]Pixel, maxY-by)
			for y := by; y < maxY; y++ {
				blockPixels[y-by] = img[y][bx:maxX]
			}

			blocks = append(blocks, Block{
				X:      bx,
				Y:      by,
				Pixels: blockPixels,
				Avg:    averageBlock(blockPixels),
			})
		}
	}
	return blocks
}

// matchBlocks associe chaque bloc cible au bloc source dont la moyenne est la plus proche (greedy).
func matchBlocks(source, target []Block) []Block {
	result := make([]Block, len(target))

	for i, tb := range target {
		best := source[0]
		bestDist := colorDistance(tb.Avg, best.Avg)

		for _, sb := range source[1:] {
			d := colorDistance(tb.Avg, sb.Avg)
			if d < bestDist {
				bestDist = d
				best = sb
			}
		}

		// placer le bloc source à la position du bloc cible
		best.X = tb.X
		best.Y = tb.Y
		result[i] = best
	}

	return result
}

// reconstructImage recolle les blocs à leur nouvelle position (en rognant si ça dépasse l'image).
func reconstructImage(blocks []Block, width, height int) [][]Pixel {
	result := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		result[y] = make([]Pixel, width)
	}

	for _, b := range blocks {
		for by := range b.Pixels {
			ay := b.Y + by
			if ay < 0 || ay >= height {
				continue
			}
			for bx := range b.Pixels[by] {
				ax := b.X + bx
				if ax < 0 || ax >= width {
					continue
				}
				result[ay][ax] = b.Pixels[by][bx]
			}
		}
	}

	return result
}

// transformToTarget : remap par blocs (approche historique, non utilisée par main.go)
// Étapes :
// 1) splitIntoBlocks sur source et cible
// 2) matchBlocks (moyenne de couleur la plus proche)
// 3) reconstructImage pour recoller les blocs à leurs nouvelles positions
// Hypothèses : images mêmes dimensions, factor > 0. Les bords sont rognés si factor ne divise pas.
func transformToTarget(
	source [][]Pixel,
	target [][]Pixel,
	width, height, factor int,
) [][]Pixel {

	srcBlocks := splitIntoBlocks(source, width, height, factor)
	tgtBlocks := splitIntoBlocks(target, width, height, factor)

	mapped := matchBlocks(srcBlocks, tgtBlocks)

	return reconstructImage(mapped, width, height)
}
