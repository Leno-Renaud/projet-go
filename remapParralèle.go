package main

import (
    "math/rand"
    "runtime"
    "sync"
)

// remapPixels (parallèle) rearranges source pixels to match the target color distribution.
// It keeps the same signature as the sequential version so existing callers need no change.
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
    rand.Shuffle(len(positions), func(i, j int) { positions[i], positions[j] = positions[j], positions[i] })

    // Feed positions into a buffered channel consumed by workers.
    posCh := make(chan [2]int, len(positions))
    for _, p := range positions {
        posCh <- p
    }
    close(posCh)

    var mu sync.Mutex // protects access to bins and popPixel
    workers := runtime.NumCPU()
    var wg sync.WaitGroup
    wg.Add(workers)

    for i := 0; i < workers; i++ {
        go func() {
            defer wg.Done()
            for pos := range posCh {
                x, y := pos[0], pos[1]
                bin := quantizePixel(target[y][x], levels)

                mu.Lock()
                p, ok := popPixel(bin, bins, levels)
                mu.Unlock()

                if !ok {
                    out[y][x] = target[y][x]
                    continue
                }
                out[y][x] = p
            }
        }()
    }

    wg.Wait()
    return out
}
