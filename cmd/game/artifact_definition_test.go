package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadArtifactDefinitionAsset(t *testing.T) {
	t.Parallel()

	assetFS := fstest.MapFS{
		"artifacts/fossil.json": &fstest.MapFile{
			Data: []byte(`{
  "name": "fossil",
  "image_path": "textures/fossil.png"
}`),
		},
	}

	definition, err := loadArtifactDefinitionAsset(assetFS, "artifacts/fossil.json")
	if err != nil {
		t.Fatalf("loadArtifactDefinitionAsset() error = %v", err)
	}

	if definition.Name != "fossil" {
		t.Fatalf("definition.Name = %q, want fossil", definition.Name)
	}
	if definition.ImagePath != "textures/fossil.png" {
		t.Fatalf("definition.ImagePath = %q, want textures/fossil.png", definition.ImagePath)
	}
}

func TestLoadArtifactDefinitionAssetValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "missing name",
			json:    `{"image_path":"textures/fossil.png"}`,
			wantErr: "artifact definition name must not be empty",
		},
		{
			name:    "missing image path",
			json:    `{"name":"fossil"}`,
			wantErr: "artifact image_path must not be empty",
		},
		{
			name:    "unknown field",
			json:    `{"name":"fossil","image_path":"textures/fossil.png","extra":true}`,
			wantErr: "decode artifact definition",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assetFS := fstest.MapFS{
				"artifacts/test.json": &fstest.MapFile{Data: []byte(test.json)},
			}

			_, err := loadArtifactDefinitionAsset(assetFS, "artifacts/test.json")
			if err == nil {
				t.Fatal("loadArtifactDefinitionAsset() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("loadArtifactDefinitionAsset() error = %q, want substring %q", err.Error(), test.wantErr)
			}
		})
	}
}
