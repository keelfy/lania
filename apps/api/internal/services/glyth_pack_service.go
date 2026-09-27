package services

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

// The glyth pack is the ItemsAdder "contents/glyth" folder: one font image per glyth prefix, so
// the in-game :glyth_<slug>: tokens render as the same icons the site shows.
const (
	glythPackNamespace = "glyth"
	// glythPackFirstSymbol is the first private-use code point the font images take, one per glyth.
	glythPackFirstSymbol rune = 0xE006
	glythPackScaleRatio       = 9
	glythPackYPosition        = 8
	glythPackImageMaxBytes    = 1 << 20
	glythPackFetchWorkers     = 8
	glythPackFetchTimeout     = 10 * time.Second
)

type GlythPackService interface {
	// BuildGlythPack returns a zip with name_prefixes.yml and textures/font/<slug>.png.
	BuildGlythPack(ctx context.Context) ([]byte, error)
}

type glythPackService struct {
	storage storage.MainStorage
	objects clients.ObjectStorage
	http    *http.Client
}

func NewGlythPackService(storage storage.MainStorage, objects clients.ObjectStorage) GlythPackService {
	return &glythPackService{
		storage: storage,
		objects: objects,
		http:    &http.Client{Timeout: glythPackFetchTimeout},
	}
}

type glythPackEntry struct {
	name  string
	slug  string
	image string
	png   []byte
	err   error
}

func (s *glythPackService) BuildGlythPack(ctx context.Context) ([]byte, error) {
	prefixes, err := s.storage.Queries().FindNamePrefixes(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get name prefixes", err)
	}

	// Special prefixes are plain text, not ItemsAdder font images, so only :glyth_<slug>: tokens count.
	entries := make([]*glythPackEntry, 0, len(prefixes))
	owners := map[string]string{}
	for _, prefix := range prefixes {
		match := glythTokenPattern.FindStringSubmatch(prefix.Metadata.Prefix)
		if match == nil {
			continue
		}
		slug := match[1]
		if owner, ok := owners[slug]; ok {
			return nil, utils.NewConflictError(fmt.Sprintf("glyths %q and %q share the token %s", owner, prefix.Name, prefix.Metadata.Prefix), nil)
		}
		owners[slug] = prefix.Name
		entries = append(entries, &glythPackEntry{name: prefix.Name, slug: slug, image: prefix.Metadata.Image})
	}
	if len(entries) == 0 {
		return nil, utils.NewNotFoundError("no glyth prefixes to export", nil)
	}
	// Symbols are handed out in this order, so it must not depend on the query order.
	sort.Slice(entries, func(i, j int) bool { return entries[i].slug < entries[j].slug })

	s.fetchImages(ctx, entries)
	var failures []string
	for _, entry := range entries {
		if entry.err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", entry.name, entry.err))
		}
	}
	// A glyth left out of the pack would show its raw token in game, so a partial pack is not useful.
	if len(failures) > 0 {
		return nil, utils.NewBadRequestError("failed to get glyth images: "+strings.Join(failures, "; "), nil)
	}

	pack, err := writeGlythPack(entries)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to build glyth pack", err)
	}
	return pack, nil
}

func (s *glythPackService) fetchImages(ctx context.Context, entries []*glythPackEntry) {
	jobs := make(chan *glythPackEntry)
	var wg sync.WaitGroup
	for range min(glythPackFetchWorkers, len(entries)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				entry.png, entry.err = s.fetchImage(ctx, entry.image)
			}
		}()
	}
	for _, entry := range entries {
		jobs <- entry
	}
	close(jobs)
	wg.Wait()
}

// fetchImage reads a glyth preview and returns it as PNG. Previews live in the project bucket
// (s3://<bucket>/<key>); a few legacy rows still hold an https URL.
func (s *glythPackService) fetchImage(ctx context.Context, location string) ([]byte, error) {
	var content []byte
	var err error
	switch {
	case location == "":
		return nil, fmt.Errorf("no image set")
	case strings.HasPrefix(location, "s3://"):
		bucket, key, _ := strings.Cut(strings.TrimPrefix(location, "s3://"), "/")
		if bucket != s.objects.Bucket() || key == "" {
			return nil, fmt.Errorf("image %s is not in bucket %s", location, s.objects.Bucket())
		}
		content, err = s.objects.GetObject(ctx, key)
	case strings.HasPrefix(location, "https://"), strings.HasPrefix(location, "http://"):
		content, err = s.download(ctx, location)
	default:
		return nil, fmt.Errorf("unsupported image location %s", location)
	}
	if err != nil {
		return nil, err
	}

	decoded, format, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("image is not readable: %w", err)
	}
	if format == "png" {
		return content, nil
	}
	// Resource pack fonts only take PNG.
	var buf bytes.Buffer
	if err := png.Encode(&buf, decoded); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *glythPackService) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image download returned %s", resp.Status)
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, glythPackImageMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(content) > glythPackImageMaxBytes {
		return nil, fmt.Errorf("image is larger than %d bytes", glythPackImageMaxBytes)
	}
	return content, nil
}

func writeGlythPack(entries []*glythPackEntry) ([]byte, error) {
	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)

	config, err := archive.Create("name_prefixes.yml")
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(config, glythPackConfig(entries)); err != nil {
		return nil, err
	}
	for _, entry := range entries {
		file, err := archive.Create("textures/font/" + entry.slug + ".png")
		if err != nil {
			return nil, err
		}
		if _, err := file.Write(entry.png); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// glythPackConfig renders the ItemsAdder font_images config. Slugs only hold [a-z0-9_], so no
// value needs escaping.
func glythPackConfig(entries []*glythPackEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "info:\n  namespace: %s\nfont_images:\n", glythPackNamespace)
	for i, entry := range entries {
		fmt.Fprintf(&b, "  %s_%s:\n", glythPackNamespace, entry.slug)
		fmt.Fprintf(&b, "    path: font/%s.png\n", entry.slug)
		fmt.Fprintf(&b, "    scale_ratio: %d\n", glythPackScaleRatio)
		fmt.Fprintf(&b, "    y_position: %d\n", glythPackYPosition)
		fmt.Fprintf(&b, "    symbol: \"%c\"\n", glythPackFirstSymbol+rune(i))
		b.WriteString("    suggest_in_command: false\n")
	}
	return b.String()
}
