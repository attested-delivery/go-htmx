//go:build e2e

// Package visual implements screenshot-based visual regression checks:
// decode two PNGs, compute per-pixel Euclidean RGB distance, fail if the
// differing-pixel fraction exceeds a threshold. Hand-rolled rather than
// a dependency — a real candidate (github.com/n7olkachev/imgdiff) exists
// but hasn't published since 2021, not worth the OSV-Scanner-monitored
// dependency surface for the ~30 lines of logic below.
package visual

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// Result is the outcome of comparing two same-sized images.
type Result struct {
	DiffPixels   int
	TotalPixels  int
	DiffFraction float64
}

// maxChannelDist is the maximum possible per-pixel Euclidean distance in
// 8-bit RGB space: math.Sqrt(3 * 255*255).
const maxChannelDist = 441.6729559300637

// pixelThreshold is how far (as a fraction of maxChannelDist) a single
// pixel's color must differ before it counts as a "differing" pixel at
// all. Below this, differences are treated as rendering noise (font
// anti-aliasing, sub-pixel rounding), not a real visual regression.
const pixelThreshold = 0.1

// Compare decodes the baseline and actual PNGs at the given paths and
// computes their per-pixel difference. If the images differ in
// dimensions, or the fraction of differing pixels exceeds threshold
// (e.g. 0.01 for 1%), it writes a diff-overlay image (differing pixels
// in red, matching pixels dimmed to gray) to diffOutPath — callers
// should treat that as a test failure and the overlay as a CI artifact
// for debugging. diffOutPath may be empty to skip writing the overlay.
func Compare(baselinePath, actualPath, diffOutPath string, threshold float64) (Result, error) {
	baseline, err := decodePNG(baselinePath)
	if err != nil {
		return Result{}, fmt.Errorf("decode baseline: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return Result{}, fmt.Errorf("decode actual: %w", err)
	}

	bb := baseline.Bounds()
	ab := actual.Bounds()
	if bb.Dx() != ab.Dx() || bb.Dy() != ab.Dy() {
		return Result{}, fmt.Errorf("dimension mismatch: baseline %dx%d, actual %dx%d", bb.Dx(), bb.Dy(), ab.Dx(), ab.Dy())
	}

	width, height := bb.Dx(), bb.Dy()
	overlay := image.NewRGBA(image.Rect(0, 0, width, height))
	diffPixels := 0

	for y := range height {
		for x := range width {
			br, bg, bl, _ := baseline.At(bb.Min.X+x, bb.Min.Y+y).RGBA()
			ar, ag, al, _ := actual.At(ab.Min.X+x, ab.Min.Y+y).RGBA()

			// RGBA() returns 16-bit-per-channel values; downshift to 8-bit.
			dr := float64(int32(br>>8) - int32(ar>>8))
			dg := float64(int32(bg>>8) - int32(ag>>8))
			db := float64(int32(bl>>8) - int32(al>>8))
			dist := math.Sqrt(dr*dr + dg*dg + db*db)

			if dist/maxChannelDist > pixelThreshold {
				diffPixels++
				overlay.Set(x, y, color.RGBA{R: 255, A: 255})
				continue
			}
			gray := uint8(ar >> 8)
			overlay.Set(x, y, color.RGBA{R: gray / 3, G: gray / 3, B: gray / 3, A: 255})
		}
	}

	total := width * height
	fraction := float64(diffPixels) / float64(total)
	result := Result{DiffPixels: diffPixels, TotalPixels: total, DiffFraction: fraction}

	if fraction <= threshold {
		return result, nil
	}

	failErr := fmt.Errorf("visual diff %.4f%% exceeds threshold %.4f%% (%d/%d pixels differ)",
		fraction*100, threshold*100, diffPixels, total)
	if diffOutPath == "" {
		return result, failErr
	}
	if err := writePNG(diffOutPath, overlay); err != nil {
		return result, fmt.Errorf("%w (also failed to write diff overlay to %s: %v)", failErr, diffOutPath, err)
	}
	return result, fmt.Errorf("%w (overlay written to %s)", failErr, diffOutPath)
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
