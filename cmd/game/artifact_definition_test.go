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
  "image_path": "textures/fossil.png",
  "width": 64,
  "height": 64,
  "value": 5
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
	if definition.Width != 64 {
		t.Fatalf("definition.Width = %d, want 64", definition.Width)
	}
	if definition.Height != 64 {
		t.Fatalf("definition.Height = %d, want 64", definition.Height)
	}
	if definition.Value != 5 {
		t.Fatalf("definition.Value = %d, want 5", definition.Value)
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
			json:    `{"name":"fossil","width":64,"height":64,"value":5}`,
			wantErr: "artifact image_path must not be empty",
		},
		{
			name:    "missing width",
			json:    `{"name":"fossil","image_path":"textures/fossil.png","height":64,"value":5}`,
			wantErr: "artifact width must be positive",
		},
		{
			name:    "missing height",
			json:    `{"name":"fossil","image_path":"textures/fossil.png","width":64,"value":5}`,
			wantErr: "artifact height must be positive",
		},
		{
			name:    "missing value",
			json:    `{"name":"fossil","image_path":"textures/fossil.png","width":64,"height":64}`,
			wantErr: "artifact value must be positive",
		},
		{
			name:    "unknown field",
			json:    `{"name":"fossil","image_path":"textures/fossil.png","width":64,"height":64,"value":5,"extra":true}`,
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
