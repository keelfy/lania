package commands

import (
	"testing"

	"github.com/google/uuid"
)

func TestUploadImageCommand_Validate(t *testing.T) {
	t.Parallel()

	content := []byte("fake-image-bytes")
	tooLarge := make([]byte, MaxUploadImageBytes[UploadImageKindGlythPreview]+1)

	tests := []struct {
		name    string
		command UploadImageCommand
		wantErr bool
	}{
		{"glyth preview", UploadImageCommand{Kind: UploadImageKindGlythPreview, Content: content, Token: ":glyth_popcat:"}, false},
		{"season screenshot", UploadImageCommand{Kind: UploadImageKindSeasonScreenshot, Content: content, SeasonID: uuid.New()}, false},
		{"season preview", UploadImageCommand{Kind: UploadImageKindSeasonPreview, Content: content}, false},
		{"unknown kind", UploadImageCommand{Kind: "banner", Content: content}, true},
		{"blank kind", UploadImageCommand{Content: content}, true},
		{"empty content", UploadImageCommand{Kind: UploadImageKindGlythPreview}, true},
		{"content over cap", UploadImageCommand{Kind: UploadImageKindGlythPreview, Content: tooLarge}, true},
		{"screenshot without season", UploadImageCommand{Kind: UploadImageKindSeasonScreenshot, Content: content}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
