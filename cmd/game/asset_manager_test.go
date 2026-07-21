package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAssetManagerLookupsReturnFalseForMissingValues(t *testing.T) {
	t.Parallel()

	assets := &AssetManager{}

	if _, ok := assets.LookupArtifactDefinition("missing"); ok {
		t.Fatal("LookupArtifactDefinition() ok = true, want false")
	}
	if _, ok := assets.LookupImage("textures/missing.png"); ok {
		t.Fatal("LookupImage() ok = true, want false")
	}
	if _, ok := assets.LookupModel("missing"); ok {
		t.Fatal("LookupModel() ok = true, want false")
	}
	if _, ok := assets.LookupShader("missing"); ok {
		t.Fatal("LookupShader() ok = true, want false")
	}
	if _, ok := assets.LookupSound("missing"); ok {
		t.Fatal("LookupSound() ok = true, want false")
	}
	if _, ok := assets.LookupTexture("missing"); ok {
		t.Fatal("LookupTexture() ok = true, want false")
	}
}

func TestMustPanicsOnMissingLookup(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("Must() did not panic, want missing value panic")
		}
	}()

	_ = Must((*ArtifactDefinition)(nil), false)
}

func TestSoundNamesReturnsSortedLoadedSoundNames(t *testing.T) {
	t.Parallel()

	assets := &AssetManager{
		sounds: map[string]rl.Music{
			"laser":   {},
			"burning": {},
			"pickup":  {},
		},
	}

	got := assets.SoundNames()
	want := []string{"burning", "laser", "pickup"}
	if len(got) != len(want) {
		t.Fatalf("len(SoundNames()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SoundNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoadImagesAndArtifactDefinitions(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	img.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})

	assetFS := fstest.MapFS{
		"textures/phone.png": &fstest.MapFile{Data: mustEncodePNGImage(t, img)},
		"artifacts/phone.json": &fstest.MapFile{Data: []byte(`{
  "name": "phone",
  "image_path": "textures/phone.png",
  "width": 64,
  "height": 64,
  "value": 10
}`)},
	}

	assets := &AssetManager{
		artifactDefinitions: make(map[string]*ArtifactDefinition),
		images:              make(map[string]image.Image),
		models:              make(map[string]*rl.Model),
		shaders:             make(map[string]rl.Shader),
		sounds:              make(map[string]rl.Music),
		textures:            make(map[string]rl.Texture2D),
		assetFS:             assetFS,
	}

	assets.loadImages()
	assets.loadArtifactDefinitions()

	if _, ok := assets.LookupImage("textures/phone.png"); !ok {
		t.Fatal("LookupImage() ok = false, want true")
	}

	definition, ok := assets.LookupArtifactDefinition("phone")
	if !ok {
		t.Fatal("LookupArtifactDefinition() ok = false, want true")
	}
	if definition.ImagePath != "textures/phone.png" {
		t.Fatalf("definition.ImagePath = %q, want textures/phone.png", definition.ImagePath)
	}
}

func TestLoadArtifactDefinitionsValidatesReferencedImages(t *testing.T) {
	t.Parallel()

	assetFS := fstest.MapFS{
		"artifacts/phone.json": &fstest.MapFile{Data: []byte(`{
  "name": "phone",
  "image_path": "textures/phone.png",
  "width": 64,
  "height": 64,
  "value": 10
}`)},
	}

	assets := &AssetManager{
		artifactDefinitions: make(map[string]*ArtifactDefinition),
		images:              make(map[string]image.Image),
		assetFS:             assetFS,
	}

	defer func() {
		if recover() == nil {
			t.Fatal("loadArtifactDefinitions() did not panic, want missing image validation")
		}
	}()

	assets.loadArtifactDefinitions()
}

func mustEncodePNG(t *testing.T, fill color.RGBA) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, fill)

	return mustEncodePNGImage(t, img)
}

func mustEncodePNGImage(t *testing.T, img image.Image) []byte {
	t.Helper()

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}

	return encoded.Bytes()
}
