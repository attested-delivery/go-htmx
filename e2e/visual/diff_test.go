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
