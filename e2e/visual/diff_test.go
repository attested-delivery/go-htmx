//go:build e2e

package visual

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeTestPNG(t *testing.T, path string, fill func(x, y int) color.Color) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := range 10 {
		for x := range 10 {
			img.Set(x, y, fill(x, y))
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}

func TestCompare_IdenticalImages(t *testing.T) {
	dir := t.TempDir()
	white := func(int, int) color.Color { return color.White }

	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, white)
	writeTestPNG(t, b, white)

	result, err := Compare(a, b, "", 0.01)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if result.DiffPixels != 0 {
		t.Errorf("DiffPixels = %d, want 0", result.DiffPixels)
	}
}

func TestCompare_EntirelyDifferentImages(t *testing.T) {
	dir := t.TempDir()
	white := func(int, int) color.Color { return color.White }
	black := func(int, int) color.Color { return color.Black }

	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, white)
	writeTestPNG(t, b, black)

	diffOut := filepath.Join(dir, "diff.png")
	result, err := Compare(a, b, diffOut, 0.01)
	if err == nil {
		t.Fatal("Compare: want an error for a fully different image, got nil")
	}
	if result.DiffPixels != 100 {
		t.Errorf("DiffPixels = %d, want 100 (all pixels)", result.DiffPixels)
	}
	if _, statErr := os.Stat(diffOut); statErr != nil {
		t.Errorf("diff overlay was not written: %v", statErr)
	}
}

func TestCompare_BelowThresholdPasses(t *testing.T) {
	dir := t.TempDir()
	// One pixel out of 100 differs -- 1%. A 5% threshold should pass.
	white := func(int, int) color.Color { return color.White }
	oneBlackPixel := func(x, y int) color.Color {
		if x == 0 && y == 0 {
			return color.Black
		}
		return color.White
	}

	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, white)
	writeTestPNG(t, b, oneBlackPixel)

	result, err := Compare(a, b, "", 0.05)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if result.DiffPixels != 1 {
		t.Errorf("DiffPixels = %d, want 1", result.DiffPixels)
	}
}

// TestCompare_OverlayGrayscaleUsesAllChannels catches a real bug found
// by Copilot review on PR #86: the overlay's dimmed (matching) pixels
// were derived from the actual pixel's red channel alone, so a bright
// green pixel dimmed to near-black instead of a mid-gray, misleadingly
// suggesting the pixel was dark when it wasn't.
func TestCompare_OverlayGrayscaleUsesAllChannels(t *testing.T) {
	dir := t.TempDir()

	// Both images are bright green (R=0) everywhere except pixel (0,0),
	// which differs (white vs. black) to force Compare's failure path —
	// Compare only writes the overlay when the diff fraction exceeds
	// threshold, so a genuine mismatch is needed to exercise it. Pixel
	// (5,5) stays matching green in both, and is what this test
	// actually inspects: a red-channel-only grayscale would dim it to
	// near-black (0/3), even though pure green's luma is ~150/255.
	green := func(x, y int) color.Color {
		if x == 0 && y == 0 {
			return color.White
		}
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}
	}
	greenWithBlackCorner := func(x, y int) color.Color {
		if x == 0 && y == 0 {
			return color.Black
		}
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}
	}

	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, green)
	writeTestPNG(t, b, greenWithBlackCorner)

	diffOut := filepath.Join(dir, "diff.png")
	if _, err := Compare(a, b, diffOut, 0.001); err == nil {
		t.Fatal("Compare: want an error (one differing pixel exceeds a 0.1% threshold), got nil")
	}

	f, err := os.Open(diffOut)
	if err != nil {
		t.Fatalf("open diff overlay: %v", err)
	}
	defer f.Close()
	overlay, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode diff overlay: %v", err)
	}

	r, g, b2, _ := overlay.At(5, 5).RGBA()
	gray := uint8(r >> 8)
	if uint8(g>>8) != gray || uint8(b2>>8) != gray {
		t.Fatalf("overlay pixel isn't neutral gray: R=%d G=%d B=%d", r>>8, g>>8, b2>>8)
	}
	// Luma of pure green (0, 255, 0) is 0.587*255 ≈ 150, dimmed by /3 ≈ 50.
	// A red-channel-only calculation would have produced 0 here.
	if gray == 0 {
		t.Error("overlay pixel for a bright green input dimmed to 0 — grayscale is still using only the red channel")
	}
}

func TestCompare_DimensionMismatch(t *testing.T) {
	dir := t.TempDir()
	white := func(int, int) color.Color { return color.White }

	a := filepath.Join(dir, "a.png")
	writeTestPNG(t, a, white)

	// A differently-sized image, written directly rather than via the
	// fixed-10x10 helper.
	b := filepath.Join(dir, "b.png")
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	f, err := os.Create(b)
	if err != nil {
		t.Fatalf("create %s: %v", b, err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", b, err)
	}
	f.Close()

	if _, err := Compare(a, b, "", 0.01); err == nil {
		t.Fatal("Compare: want an error for mismatched dimensions, got nil")
	}
}
