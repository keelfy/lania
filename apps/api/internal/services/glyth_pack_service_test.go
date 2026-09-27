package services

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

func glythPrefix(name, token, image string) *domain.NamePrefix {
	return &domain.NamePrefix{
		ID:       uuid.New(),
		Name:     name,
		Metadata: domain.NamePrefixMetadata{Prefix: token, Image: image},
	}
}

func readZip(t *testing.T, content []byte) map[string][]byte {
	t.Helper()
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("pack is not a zip: %v", err)
	}
	files := map[string][]byte{}
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open %s: %v", file.Name, err)
		}
		body, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatalf("failed to read %s: %v", file.Name, err)
		}
		files[file.Name] = body
	}
	return files
}

func TestGlythPackService_BuildGlythPack(t *testing.T) {
	t.Parallel()

	fox := encodePNG(t, 32, 32)
	jokerge := encodePNG(t, 16, 16)
	objects := &fakeObjectStorage{bucket: "lania-web", objects: map[string][]byte{
		"glyth_preview/fox-aaaa1111.png":     fox,
		"glyth_preview/jokerge-bbbb2222.png": jokerge,
		"glyth_preview/arcanemist.jpg":       encodeJPEG(t, 32, 32),
	}}
	queries := &fakeCatalogQueries{prefixes: []*domain.NamePrefix{
		glythPrefix("Jokerge", ":glyth_jokerge:", "s3://lania-web/glyth_preview/jokerge-bbbb2222.png"),
		glythPrefix("Fox", ":glyth_fox:", "s3://lania-web/glyth_preview/fox-aaaa1111.png"),
		glythPrefix("Admin", "[Admin]", ""),
		glythPrefix("Arcanemist", ":glyth_arcanemist:", "s3://lania-web/glyth_preview/arcanemist.jpg"),
	}}
	service := NewGlythPackService(&fakeCatalogStorage{queries: queries}, objects)

	pack, err := service.BuildGlythPack(context.Background())
	if err != nil {
		t.Fatalf("BuildGlythPack() error = %v", err)
	}
	files := readZip(t, pack)

	wantConfig := "info:\n  namespace: glyth\nfont_images:\n" +
		"  glyth_arcanemist:\n    path: font/arcanemist.png\n    scale_ratio: 9\n    y_position: 8\n    symbol: \"\"\n    suggest_in_command: false\n" +
		"  glyth_fox:\n    path: font/fox.png\n    scale_ratio: 9\n    y_position: 8\n    symbol: \"\"\n    suggest_in_command: false\n" +
		"  glyth_jokerge:\n    path: font/jokerge.png\n    scale_ratio: 9\n    y_position: 8\n    symbol: \"\"\n    suggest_in_command: false\n"
	if got := string(files["name_prefixes.yml"]); got != wantConfig {
		t.Errorf("name_prefixes.yml =\n%s\nwant\n%s", got, wantConfig)
	}
	if !bytes.Equal(files["textures/font/fox.png"], fox) {
		t.Errorf("fox.png is not the stored preview")
	}
	if !bytes.Equal(files["textures/font/jokerge.png"], jokerge) {
		t.Errorf("jokerge.png is not the stored preview")
	}
	if !bytes.HasPrefix(files["textures/font/arcanemist.png"], []byte("\x89PNG")) {
		t.Errorf("arcanemist.png was not converted to PNG")
	}
	if len(files) != 4 {
		t.Errorf("pack has %d files, want 4", len(files))
	}
}

func TestGlythPackService_BuildGlythPackErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prefixes   []*domain.NamePrefix
		wantStatus int
		wantText   string
	}{
		{
			name:       "no glyths",
			prefixes:   []*domain.NamePrefix{glythPrefix("Admin", "[Admin]", "")},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "missing image names the glyth",
			prefixes: []*domain.NamePrefix{
				glythPrefix("Fox", ":glyth_fox:", "s3://lania-web/glyth_preview/missing.png"),
				glythPrefix("Cat", ":glyth_cat:", ""),
			},
			wantStatus: http.StatusBadRequest,
			wantText:   "Cat: no image set",
		},
		{
			name:       "image in another bucket",
			prefixes:   []*domain.NamePrefix{glythPrefix("Fox", ":glyth_fox:", "s3://other/fox.png")},
			wantStatus: http.StatusBadRequest,
			wantText:   "Fox:",
		},
		{
			name: "duplicate token",
			prefixes: []*domain.NamePrefix{
				glythPrefix("Fox", ":glyth_fox:", ""),
				glythPrefix("Fox2", ":glyth_fox:", ""),
			},
			wantStatus: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			queries := &fakeCatalogQueries{prefixes: tt.prefixes}
			objects := &fakeObjectStorage{bucket: "lania-web", objects: map[string][]byte{}}
			service := NewGlythPackService(&fakeCatalogStorage{queries: queries}, objects)

			_, err := service.BuildGlythPack(context.Background())
			if err == nil {
				t.Fatal("BuildGlythPack() error = nil")
			}
			if status := utils.MapCustomErrorToHttpStatus(err); status != tt.wantStatus {
				t.Errorf("status = %d, want %d (%v)", status, tt.wantStatus, err)
			}
			if !strings.Contains(err.Error(), tt.wantText) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantText)
			}
		})
	}
}
