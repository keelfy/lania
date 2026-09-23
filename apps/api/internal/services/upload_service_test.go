package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
)

type fakeObjectStorage struct {
	bucket   string
	putKey   string
	putBody  []byte
	putCalls int
}

func (s *fakeObjectStorage) Bucket() string { return s.bucket }

func (s *fakeObjectStorage) PutObject(_ context.Context, key string, body []byte, _ string) error {
	s.putCalls++
	s.putKey = key
	s.putBody = body
	return nil
}

func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode fixture PNG: %v", err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("failed to encode fixture JPEG: %v", err)
	}
	return buf.Bytes()
}

func TestUploadService_UploadImage(t *testing.T) {
	t.Parallel()

	seasonID := uuid.New()

	tests := []struct {
		name    string
		cmd     commands.UploadImageCommand
		wantErr bool
		wantKey string
	}{
		{
			name:    "glyth key derived from token",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: encodePNG(t, 32, 32), Token: ":glyth_popcat:"},
			wantKey: "glyth_preview/popcat-",
		},
		{
			name:    "glyth falls back to name when token is malformed",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: encodePNG(t, 32, 32), Token: "not-a-token", Name: "Zip Zap!"},
			wantKey: "glyth_preview/zip_zap-",
		},
		{
			name:    "glyth rejects jpeg",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: encodeJPEG(t, 32, 32), Token: ":glyth_popcat:"},
			wantErr: true,
		},
		{
			name:    "glyth rejects non-square",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: encodePNG(t, 32, 16), Token: ":glyth_popcat:"},
			wantErr: true,
		},
		{
			name:    "screenshot below minimum size is rejected",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindSeasonScreenshot, Content: encodePNG(t, 100, 100), SeasonID: seasonID},
			wantErr: true,
		},
		{
			name:    "screenshot key is scoped to the season",
			cmd:     commands.UploadImageCommand{Kind: commands.UploadImageKindSeasonScreenshot, Content: encodeJPEG(t, 1280, 720), SeasonID: seasonID},
			wantKey: "season_screenshots/" + seasonID.String() + "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := &fakeObjectStorage{bucket: "lania-web"}
			service := NewUploadService(storage)

			result, err := service.UploadImage(context.Background(), &tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UploadImage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if storage.putCalls != 0 {
					t.Errorf("PutObject called %d times on a rejected upload", storage.putCalls)
				}
				return
			}
			if storage.putCalls != 1 {
				t.Fatalf("PutObject called %d times, want 1", storage.putCalls)
			}
			if got := storage.putKey; len(got) < len(tt.wantKey) || got[:len(tt.wantKey)] != tt.wantKey {
				t.Errorf("key = %q, want prefix %q", got, tt.wantKey)
			}
			wantLocation := "s3://lania-web/" + storage.putKey
			if result.Location != wantLocation {
				t.Errorf("Location = %q, want %q", result.Location, wantLocation)
			}
		})
	}
}

func TestUploadService_UploadImage_IdenticalBytesProduceTheSameKey(t *testing.T) {
	t.Parallel()

	content := encodePNG(t, 32, 32)
	storage := &fakeObjectStorage{bucket: "lania-web"}
	service := NewUploadService(storage)

	first, err := service.UploadImage(context.Background(), &commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: content, Token: ":glyth_popcat:"})
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	second, err := service.UploadImage(context.Background(), &commands.UploadImageCommand{Kind: commands.UploadImageKindGlythPreview, Content: content, Token: ":glyth_popcat:"})
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if first.Location != second.Location {
		t.Errorf("Location = %q, then %q, want the same key for identical bytes", first.Location, second.Location)
	}
}
