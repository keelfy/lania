package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg" // registers the JPEG decoder with image.DecodeConfig
	_ "image/png"  // registers the PNG decoder with image.DecodeConfig
	"regexp"
	"strings"

	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

var (
	glythTokenPattern  = regexp.MustCompile(`^:glyth_([a-z0-9_]+):$`)
	nonSlugCharPattern = regexp.MustCompile(`[^a-z0-9]+`)
)

// uploadImageRules bounds what a kind accepts. Checked after decoding, so the real format and
// dimensions are used, never the browser-supplied Content-Type.
type uploadImageRules struct {
	formats       map[string]bool
	minWidth      int
	minHeight     int
	maxWidth      int
	maxHeight     int
	requireSquare bool
}

var uploadRulesByKind = map[commands.UploadImageKind]uploadImageRules{
	// PNG only, square, sized for a chat-line icon. imgproxy never enlarges, so a source smaller
	// than the rendered size would stay small, and a non-square source would be cropped by the
	// callers that render a fixed square.
	commands.UploadImageKindGlythPreview: {
		formats:       map[string]bool{"png": true},
		minWidth:      16,
		minHeight:     16,
		maxWidth:      256,
		maxHeight:     256,
		requireSquare: true,
	},
	commands.UploadImageKindSeasonScreenshot: {
		formats:  map[string]bool{"png": true, "jpeg": true},
		minWidth: 640, minHeight: 360,
		maxWidth: 8192, maxHeight: 8192,
	},
	commands.UploadImageKindSeasonPreview: {
		formats:  map[string]bool{"png": true, "jpeg": true},
		minWidth: 640, minHeight: 360,
		maxWidth: 8192, maxHeight: 8192,
	},
}

type UploadService interface {
	UploadImage(ctx context.Context, cmd *commands.UploadImageCommand) (*domain.UploadedImage, error)
}

type uploadService struct{ storage clients.ObjectStorage }

func NewUploadService(storage clients.ObjectStorage) UploadService {
	return &uploadService{storage: storage}
}

func (s *uploadService) UploadImage(ctx context.Context, cmd *commands.UploadImageCommand) (*domain.UploadedImage, error) {
	rules, ok := uploadRulesByKind[cmd.Kind]
	if !ok {
		return nil, utils.NewBadRequestError(fmt.Sprintf("unsupported upload kind %q", cmd.Kind), nil)
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(cmd.Content))
	if err != nil {
		return nil, utils.NewBadRequestError("file is not a readable image", err)
	}
	if !rules.formats[format] {
		return nil, utils.NewBadRequestError(fmt.Sprintf("%s images are not accepted for this upload", format), nil)
	}
	if config.Width < rules.minWidth || config.Height < rules.minHeight || config.Width > rules.maxWidth || config.Height > rules.maxHeight {
		return nil, utils.NewBadRequestError(fmt.Sprintf("image must be between %dx%d and %dx%d", rules.minWidth, rules.minHeight, rules.maxWidth, rules.maxHeight), nil)
	}
	if rules.requireSquare && config.Width != config.Height {
		return nil, utils.NewBadRequestError("image must be square", nil)
	}

	hash := sha256.Sum256(cmd.Content)
	hash8 := hex.EncodeToString(hash[:])[:8]
	contentType := "image/" + format
	ext := format
	if format == "jpeg" {
		ext = "jpg"
	}

	var key string
	switch cmd.Kind {
	case commands.UploadImageKindGlythPreview:
		key = fmt.Sprintf("glyth_preview/%s-%s.png", glythPreviewSlug(cmd.Token, cmd.Name), hash8)
	case commands.UploadImageKindSeasonScreenshot:
		key = fmt.Sprintf("season_screenshots/%s/%s.%s", cmd.SeasonID, hash8, ext)
	case commands.UploadImageKindSeasonPreview:
		key = fmt.Sprintf("season_previews/%s.%s", hash8, ext)
	}

	if err := s.storage.PutObject(ctx, key, cmd.Content, contentType); err != nil {
		return nil, utils.NewInternalServerError("failed to upload image", err)
	}

	return &domain.UploadedImage{
		Location: fmt.Sprintf("s3://%s/%s", s.storage.Bucket(), key),
		Width:    config.Width,
		Height:   config.Height,
	}, nil
}

// glythPreviewSlug derives the readable part of a glyth preview key from the in-game token
// (":glyth_popcat:" -> "popcat"), matching the convention the glyth-preview-to-S3 migration
// established. A token that doesn't match falls back to a slugified cosmetic name.
func glythPreviewSlug(token, name string) string {
	if match := glythTokenPattern.FindStringSubmatch(token); match != nil {
		return match[1]
	}
	return slugify(name)
}

func slugify(value string) string {
	slug := strings.Trim(nonSlugCharPattern.ReplaceAllString(strings.ToLower(value), "_"), "_")
	if slug == "" {
		return "image"
	}
	return slug
}
