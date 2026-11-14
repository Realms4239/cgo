package frontend

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestEmbeddedAssetsPresent(t *testing.T) {
	js, err := fs.Glob(FS, "assets/*.js")
	if err != nil || len(js) == 0 {
		t.Skip("frontend non construit — make frontend")
	}
	if err := ValidateEmbeddedAssets(); err != nil {
		t.Fatalf("embedded SPA missing: %v", err)
	}
}

func TestValidateAssets(t *testing.T) {
	index := []byte("<!doctype html><div id=\"root\"></div>")
	tests := []struct {
		name    string
		fsys    fstest.MapFS
		wantErr bool
	}{
		{"empty", fstest.MapFS{}, true},
		{"index only", fstest.MapFS{
			"index.html": {Data: index},
		}, true},
		{"index + bundle", fstest.MapFS{
			"index.html":    {Data: index},
			"assets/app.js": {Data: []byte("console.log('spa')")},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAssets(tt.fsys)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAssets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
