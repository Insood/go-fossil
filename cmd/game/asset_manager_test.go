package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
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
	if _, ok := assets.LookupAnimation("missing"); ok {
		t.Fatal("LookupAnimation() ok = true, want false")
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

func TestDecodeGIFAnimationAssetReturnsCompositedFrames(t *testing.T) {
	t.Parallel()

	assetFS := fstest.MapFS{
		"animations/pulse.gif": &fstest.MapFile{Data: mustEncodeGIFAnimation(t)},
	}

	frames, width, height := decodeGIFAnimationAsset(assetFS, "animations/pulse.gif")
	if got, want := width, 2; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := height, 2; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := len(frames), 2; got != want {
		t.Fatalf("frame count = %d, want %d", got, want)
	}
	if got, want := frames[0].duration, float32(0.05); got != want {
		t.Fatalf("frame 0 duration = %.2f, want %.2f", got, want)
	}
	if got, want := frames[1].duration, float32(0.1); got != want {
		t.Fatalf("frame 1 duration = %.2f, want %.2f", got, want)
	}

	if got := color.RGBAModel.Convert(frames[0].image.At(0, 0)).(color.RGBA); got.R != 255 || got.A != 255 {
		t.Fatalf("frame 0 pixel (0,0) = %#v, want opaque red", got)
	}
	if got := color.RGBAModel.Convert(frames[1].image.At(0, 0)).(color.RGBA); got.R != 255 || got.A != 255 {
		t.Fatalf("frame 1 pixel (0,0) = %#v, want red carried from previous frame", got)
	}
	if got := color.RGBAModel.Convert(frames[1].image.At(1, 1)).(color.RGBA); got.B != 255 || got.A != 255 {
		t.Fatalf("frame 1 pixel (1,1) = %#v, want opaque blue", got)
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

func mustEncodeGIFAnimation(t *testing.T) []byte {
	t.Helper()

	palette := color.Palette{
		color.RGBA{},
		color.RGBA{R: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	}
	first := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	first.SetColorIndex(0, 0, 1)
	second := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	second.SetColorIndex(1, 1, 2)

	animation := gif.GIF{
		Image: []*image.Paletted{first, second},
		Delay: []int{5, 10},
		Config: image.Config{
			ColorModel: palette,
			Width:      2,
			Height:     2,
		},
	}

	var encoded bytes.Buffer
	if err := gif.EncodeAll(&encoded, &animation); err != nil {
		t.Fatalf("gif.EncodeAll() error = %v", err)
	}

	return encoded.Bytes()
}
