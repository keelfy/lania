package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/domain"
	"github.com/lania-smp/shell/internal/logger"
	"github.com/lania-smp/shell/internal/storage"
)

// ErrInvalidSkin means the skin is neither a link SkinsRestorer accepts nor a nickname with its account.
var ErrInvalidSkin = errors.New("skin is neither a link nor a nickname")

// ErrSkinNotApplied means SkinsRestorer did not apply the skin in time. It may still apply it later, or it may
// have refused it.
var ErrSkinNotApplied = errors.New("skin is not applied in time")

// Skin variants as the SkinsRestorer command takes them. Empty lets it guess the model.
const (
	SkinVariantAny     = ""
	SkinVariantClassic = "classic"
	SkinVariantSlim    = "slim"
)

// skinURLPattern keeps quotes, spaces and backslashes out of the console input. SkinsRestorer stores a link in an
// ASCII column of 266 characters.
var skinURLPattern = regexp.MustCompile(`^https?://[!#-\[\]-~]{1,258}$`)

const (
	defaultSkinApplyTimeout = 15 * time.Second
	defaultSkinPollInterval = 500 * time.Millisecond
)

type SkinService interface {
	// GetPlayerSkins returns the skin of every player that chose one with SkinsRestorer.
	GetPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.PlayerSkin, error)
	// SetPlayerSkin gives the player the skin of a link, or of the licensed account mojangUUID by its nickname,
	// and waits until SkinsRestorer applies it. It fails with ErrSkinNotApplied when that takes too long.
	SetPlayerSkin(ctx context.Context, mcUUID uuid.UUID, skin, variant string, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error)
	// ClearPlayerSkin takes the SkinsRestorer skin off the player and waits until SkinsRestorer does it. It fails
	// with ErrSkinNotApplied when that takes too long.
	ClearPlayerSkin(ctx context.Context, mcUUID uuid.UUID) error
}

type skinService struct {
	skinsRestorerStorage storage.SkinsRestorerStorage
	console              clients.Console
	applyTimeout         time.Duration
	pollInterval         time.Duration
}

func NewSkinService(skinsRestorerStorage storage.SkinsRestorerStorage, console clients.Console) SkinService {
	return &skinService{
		skinsRestorerStorage: skinsRestorerStorage,
		console:              console,
		applyTimeout:         defaultSkinApplyTimeout,
		pollInterval:         defaultSkinPollInterval,
	}
}

func (s *skinService) GetPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.PlayerSkin, error) {
	records, err := s.skinsRestorerStorage.FindPlayerSkins(ctx, mcUUIDs)
	if err != nil {
		return nil, err
	}

	skins := make(map[uuid.UUID]*domain.PlayerSkin, len(records))
	for mcUUID, record := range records {
		skin, err := decodeSkin(record)
		if err != nil {
			logger.Warnf(ctx, "failed to decode the skin of %s: %v", mcUUID, err)
			continue
		}
		if skin != nil {
			skins[mcUUID] = skin
		}
	}
	return skins, nil
}

// skinWanted tells whether the player wears the skin asked for.
type skinWanted func(record *storage.SkinRecord) bool

func (s *skinService) SetPlayerSkin(ctx context.Context, mcUUID uuid.UUID, skin, variant string, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error) {
	if variant != SkinVariantAny && variant != SkinVariantClassic && variant != SkinVariantSlim {
		return nil, fmt.Errorf("%w: unknown variant %q", ErrInvalidSkin, variant)
	}

	var wanted skinWanted
	switch {
	case skinURLPattern.MatchString(skin):
		wanted = func(record *storage.SkinRecord) bool {
			return record != nil && record.Type == storage.SkinTypeURL && record.Identifier == skin && record.Value != nil &&
				(variant == SkinVariantAny || record.Variant != nil && strings.EqualFold(*record.Variant, variant))
		}
	case usernamePattern.MatchString(skin) && mojangUUID != nil:
		// The account has its own model.
		variant = SkinVariantAny
		wanted = func(record *storage.SkinRecord) bool {
			if record == nil || record.Type != storage.SkinTypePlayer {
				return false
			}
			identifier, err := uuid.Parse(record.Identifier)
			return err == nil && identifier == *mojangUUID
		}
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidSkin, skin)
	}

	command := strings.TrimSpace(strings.NewReplacer("{skin}", skin, "{uuid}", mcUUID.String(), "{variant}", variant).Replace(config.GetSkinSetCommand()))
	record, err := s.apply(ctx, mcUUID, command, wanted)
	if err != nil {
		return nil, err
	}
	decoded, err := decodeSkin(record)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return &domain.PlayerSkin{}, nil
	}
	return decoded, nil
}

func (s *skinService) ClearPlayerSkin(ctx context.Context, mcUUID uuid.UUID) error {
	command := strings.ReplaceAll(config.GetSkinClearCommand(), "{uuid}", mcUUID.String())
	_, err := s.apply(ctx, mcUUID, command, func(record *storage.SkinRecord) bool { return record == nil })
	return err
}

// apply runs the command unless the player already wears the skin, then waits until the player does.
// SkinsRestorer runs console commands in background, so the console output tells nothing.
func (s *skinService) apply(ctx context.Context, mcUUID uuid.UUID, command string, wanted skinWanted) (*storage.SkinRecord, error) {
	record, err := s.findSkin(ctx, mcUUID)
	if err != nil {
		return nil, err
	}
	if wanted(record) {
		return record, nil
	}

	if _, err := s.console.Execute(ctx, command); err != nil {
		return nil, fmt.Errorf("failed to run %q: %w", command, err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.applyTimeout)
	defer cancel()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ErrSkinNotApplied
		case <-ticker.C:
		}

		record, err := s.findSkin(ctx, mcUUID)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, ErrSkinNotApplied
		} else if err != nil {
			return nil, err
		}
		if wanted(record) {
			return record, nil
		}
	}
}

func (s *skinService) findSkin(ctx context.Context, mcUUID uuid.UUID) (*storage.SkinRecord, error) {
	records, err := s.skinsRestorerStorage.FindPlayerSkins(ctx, uuid.UUIDs{mcUUID})
	if err != nil {
		return nil, err
	}
	return records[mcUUID], nil
}

// texturesProperty is the decoded "textures" property of a Minecraft profile.
type texturesProperty struct {
	Textures struct {
		Skin *struct {
			URL      string `json:"url"`
			Metadata *struct {
				Model string `json:"model"`
			} `json:"metadata"`
		} `json:"SKIN"`
	} `json:"textures"`
}

// decodeSkin reads the texture and the model from the textures property. A copied licensed skin SkinsRestorer has
// not fetched yet gives the account only. A record with no data gives nil.
func decodeSkin(record *storage.SkinRecord) (*domain.PlayerSkin, error) {
	if record == nil {
		return nil, nil
	}

	skin := &domain.PlayerSkin{}
	if record.Type == storage.SkinTypePlayer {
		if mojangUUID, err := uuid.Parse(record.Identifier); err == nil {
			skin.MojangUUID = &mojangUUID
		}
	}
	if record.Value == nil {
		if skin.MojangUUID == nil {
			return nil, nil
		}
		return skin, nil
	}

	raw, err := base64.StdEncoding.DecodeString(*record.Value)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(*record.Value)
	}
	if err != nil {
		return nil, fmt.Errorf("textures property is not base64: %w", err)
	}
	var property texturesProperty
	if err := json.Unmarshal(raw, &property); err != nil {
		return nil, fmt.Errorf("textures property is not json: %w", err)
	}
	if property.Textures.Skin == nil || property.Textures.Skin.URL == "" {
		return nil, errors.New("textures property has no skin")
	}

	skin.TextureURL = strings.Replace(property.Textures.Skin.URL, "http://", "https://", 1)
	skin.Slim = property.Textures.Skin.Metadata != nil && property.Textures.Skin.Metadata.Model == "slim"
	return skin, nil
}
