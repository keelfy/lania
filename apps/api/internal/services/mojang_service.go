package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
)

type MojangService interface {
	GetMojangUUID(ctx context.Context, username string) (uuid.UUID, error)
	IsPlayerModelSlim(ctx context.Context, uuid uuid.UUID) (bool, error)
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

func (s *mojangService) GetMojangUUID(ctx context.Context, username string) (uuid.UUID, error) {
	cacheKey := fmt.Sprintf("mojang_uuid:%s", username)
	cacheValue, err := s.cache.GetKey(ctx, cacheKey)
	if err == nil {
		mojangUUID, err := uuid.Parse(cacheValue)
		if err == nil {
			return mojangUUID, nil
		}

		logger.Debugf(ctx, "[MOJANG] Cache hit for %s but failed to parse UUID: %v", username, err)
	}

	url := fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username)
	response, err := http.Get(url)
	if err != nil {
		return uuid.Nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		logger.Debugf(ctx, "[MOJANG] Mojang UUID not found: %s", response.Status)
		err = s.cache.SetKey(ctx, cacheKey, uuid.Nil.String(), 24*time.Hour)
		if err != nil {
			logger.Debugf(ctx, "[MOJANG] Error setting cache for %s: %v", username, err)
		}
		return uuid.Nil, fmt.Errorf("mojang uuid not found: %s", response.Status)
	} else if response.StatusCode != http.StatusOK {
		logger.Errorf(ctx, "[MOJANG] Failed to get mojang uuid: %s", response.Status)
		return uuid.Nil, fmt.Errorf("failed to get mojang uuid: %s", response.Status)
	}

	var mojangResponse mojangUUIDResponse
	err = json.NewDecoder(response.Body).Decode(&mojangResponse)
	if err != nil {
		return uuid.Nil, err
	}

	mojangUUID, err := uuid.Parse(mojangResponse.ID)
	if err != nil {
		return uuid.Nil, err
	}

	err = s.cache.SetKey(ctx, cacheKey, mojangUUID.String(), 24*time.Hour)
	if err != nil {
		logger.Debugf(ctx, "[MOJANG] Error setting cache for %s: %v", username, err)
	}

	return mojangUUID, nil
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
