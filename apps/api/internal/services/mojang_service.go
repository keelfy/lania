package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

const (
	mojangBulkLookupURL  = "https://api.minecraftservices.com/minecraft/profile/lookup/bulk/byname"
	mojangBulkLookupSize = 10 // max names per bulk request

	mojangSyncInterval         = 30 * time.Second
	mojangSyncBatchesPerRun    = 5
	mojangSyncBatchDelay       = time.Second
	mojangSyncRateLimitBackoff = 10 * time.Minute
	mojangNotFoundRecheckAfter = 24 * time.Hour
)

var (
	errMojangRateLimited = errors.New("mojang rate limit exceeded")
	mojangUsernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]{3,16}$`)
	mojangHTTPClient     = &http.Client{Timeout: 10 * time.Second}
)

type MojangService interface {
	// GetMojangUUIDsByMinecraftUUIDs reads Mojang UUIDs synced in background.
	// Profiles without a Mojang account are absent from the result.
	GetMojangUUIDsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error)
	IsPlayerModelSlim(ctx context.Context, uuid uuid.UUID) (bool, error)
	// ResolveGameUUID returns the UUID the server gives the nickname: the Mojang UUID when Mojang has an account
	// with it, the offline UUID otherwise. mojangUUID is nil for a free nickname. It asks Mojang every time and
	// fails with a service unavailable error when Mojang cannot answer, so the UUID is never guessed.
	ResolveGameUUID(ctx context.Context, username string) (gameUUID uuid.UUID, mojangUUID *uuid.UUID, err error)
	// RunProfileSync looks up Mojang UUIDs of profiles until ctx is done.
	RunProfileSync(ctx context.Context)
}

type mojangService struct {
	storage storage.MainStorage
	cache   storage.CacheStorage
}

func NewMojangService(
	storage storage.MainStorage,
	cache storage.CacheStorage,
) MojangService {
	return &mojangService{
		storage: storage,
		cache:   cache,
	}
}

type mojangUUIDResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *mojangService) GetMojangUUIDsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	mojangUUIDs, err := s.storage.Queries().FindProfileMojangUUIDsByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get mojang uuids by minecraft uuids", err)
	}
	return mojangUUIDs, nil
}

func (s *mojangService) ResolveGameUUID(ctx context.Context, username string) (uuid.UUID, *uuid.UUID, error) {
	offlineUUID, err := utils.GetOfflinePlayerUUID(username)
	if err != nil {
		return uuid.Nil, nil, err
	}
	// Mojang rejects such usernames, and no account can have one.
	if !mojangUsernameRegexp.MatchString(username) {
		return offlineUUID, nil, nil
	}

	found, err := s.lookupUUIDsByUsernames(ctx, []string{username})
	if err != nil {
		return uuid.Nil, nil, utils.NewServiceUnavailableError("failed to check the nickname on Mojang, try again later", err)
	}
	if mojangUUID, ok := found[strings.ToLower(username)]; ok {
		return mojangUUID, &mojangUUID, nil
	}
	return offlineUUID, nil, nil
}

func (s *mojangService) RunProfileSync(ctx context.Context) {
	wait := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		wait = mojangSyncInterval
		err := s.syncProfiles(ctx)
		switch {
		case err == nil, errors.Is(err, context.Canceled):
		case errors.Is(err, errMojangRateLimited):
			logger.Warnf(ctx, "[MOJANG] Rate limited, sync paused for %s", mojangSyncRateLimitBackoff)
			wait = mojangSyncRateLimitBackoff
		default:
			logger.Errorf(ctx, "[MOJANG] Failed to sync mojang uuids: %v", err)
		}
	}
}

func (s *mojangService) syncProfiles(ctx context.Context) error {
	targets, err := s.storage.Queries().FindMojangLookupTargets(
		ctx,
		time.Now().Add(-mojangNotFoundRecheckAfter),
		mojangBulkLookupSize*mojangSyncBatchesPerRun,
	)
	if err != nil {
		return err
	}

	for start := 0; start < len(targets); start += mojangBulkLookupSize {
		if start > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(mojangSyncBatchDelay):
			}
		}

		batch := targets[start:min(start+mojangBulkLookupSize, len(targets))]
		if err := s.syncBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func (s *mojangService) syncBatch(ctx context.Context, batch []*domain.MojangLookupTarget) error {
	usernames := make([]string, 0, len(batch))
	for _, target := range batch {
		// Mojang rejects the whole request on an invalid username, so such profiles are stored as not found.
		if mojangUsernameRegexp.MatchString(target.MinecraftUsername) {
			usernames = append(usernames, target.MinecraftUsername)
		}
	}

	found := map[string]uuid.UUID{}
	if len(usernames) > 0 {
		var err error
		found, err = s.lookupUUIDsByUsernames(ctx, usernames)
		if err != nil {
			return err
		}
	}

	for _, target := range batch {
		var mojangUUID *uuid.UUID
		if id, ok := found[strings.ToLower(target.MinecraftUsername)]; ok {
			mojangUUID = &id
		}

		if err := s.storage.Queries().UpsertProfileMojangUUID(ctx, target.MinecraftUUID, mojangUUID); err != nil {
			return err
		}
		// The nickname was free and somebody bought it since: the player can no longer join with it.
		// Moving the profile to the buyer's UUID would give it away, so an admin sorts it out.
		if mojangUUID != nil && target.CheckedBefore {
			logger.Warnf(ctx, "[MOJANG] Nickname %s of profile %s became a licensed account, marked as premium conflict", target.MinecraftUsername, target.MinecraftUUID)
			if err := s.storage.Queries().SetProfilePremiumConflict(ctx, target.MinecraftUUID); err != nil {
				return err
			}
		}
	}
	return nil
}

// lookupUUIDsByUsernames returns Mojang UUIDs by lowercase usernames. Unknown usernames are absent.
func (s *mojangService) lookupUUIDsByUsernames(ctx context.Context, usernames []string) (map[string]uuid.UUID, error) {
	body, err := json.Marshal(usernames)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, mojangBulkLookupURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := mojangHTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		return nil, errMojangRateLimited
	} else if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup mojang uuids: %s", response.Status)
	}

	var profiles []mojangUUIDResponse
	if err := json.NewDecoder(response.Body).Decode(&profiles); err != nil {
		return nil, err
	}

	found := make(map[string]uuid.UUID, len(profiles))
	for _, profile := range profiles {
		mojangUUID, err := uuid.Parse(profile.ID)
		if err != nil {
			return nil, err
		}
		found[strings.ToLower(profile.Name)] = mojangUUID
	}
	return found, nil
}

type mojangPlayerProfileProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mojangPlayerProfileResponse struct {
	ID         string                        `json:"id"`
	Properties []mojangPlayerProfileProperty `json:"properties"`
}

type texturesResponse struct {
	Timestamp   int64  `json:"timestamp"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	Textures    struct {
		SKIN struct {
			URL      string `json:"url"`
			Metadata struct {
				Model string `json:"model"`
			} `json:"metadata"`
		} `json:"SKIN"`
		CAPE struct {
			URL string `json:"url"`
		} `json:"CAPE"`
	} `json:"textures"`
}

func (s *mojangService) IsPlayerModelSlim(ctx context.Context, uuid uuid.UUID) (bool, error) {
	cacheKey := fmt.Sprintf("mojang_player_model_slim:%s", uuid)
	cacheValue, err := s.cache.GetBoolean(ctx, cacheKey)
	if err == nil {
		return cacheValue, nil
	} else {
		logger.Debugf(ctx, "[MOJANG] Error getting cache for %s: %v", uuid, err)
	}

	url := fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/profile/%s", uuid)
	response, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get player model slim: %s", response.Status)
	}

	var profile mojangPlayerProfileResponse
	err = json.NewDecoder(response.Body).Decode(&profile)
	if err != nil {
		return false, err
	}

	var textures texturesResponse
	for _, property := range profile.Properties {
		if property.Name == "textures" {
			base64Decoded, err := base64.StdEncoding.DecodeString(property.Value)
			if err != nil {
				return false, err
			}

			err = json.Unmarshal(base64Decoded, &textures)
			if err != nil {
				return false, err
			}
			break
		}
	}

	isSlim := strings.ToLower(textures.Textures.SKIN.Metadata.Model) == "slim"
	err = s.cache.SetKey(ctx, cacheKey, isSlim, 24*time.Hour)
	if err != nil {
		logger.Debugf(ctx, "[MOJANG] Error setting cache for %s: %v", uuid.String(), err)
	}
	return isSlim, nil
}
