package services

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/storage"
)

// fakeSkinsRestorerStorage answers with the next state on every read and keeps the last one.
type fakeSkinsRestorerStorage struct {
	mcUUID uuid.UUID
	states []*storage.SkinRecord
	reads  int
}

func (s *fakeSkinsRestorerStorage) FindPlayerSkins(_ context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*storage.SkinRecord, error) {
	state := s.states[min(s.reads, len(s.states)-1)]
	s.reads++
	records := make(map[uuid.UUID]*storage.SkinRecord)
	for _, mcUUID := range mcUUIDs {
		if mcUUID == s.mcUUID && state != nil {
			records[mcUUID] = state
		}
	}
	return records, nil
}

func texturesValue(json string) *string {
	value := base64.StdEncoding.EncodeToString([]byte(json))
	return &value
}

func ptr[T any](value T) *T {
	return &value
}

func newTestSkinService(storage storage.SkinsRestorerStorage, console *outputConsole) *skinService {
	service := NewSkinService(storage, console).(*skinService)
	service.applyTimeout = 50 * time.Millisecond
	service.pollInterval = time.Millisecond
	return service
}

func TestSkinServiceGetPlayerSkins(t *testing.T) {
	slim, classic, licensed, fetched, broken := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	account := uuid.New()
	records := map[uuid.UUID]*storage.SkinRecord{
		slim: {Type: storage.SkinTypeURL, Identifier: "https://lania.example/a.png", Variant: ptr("SLIM"),
			Value: texturesValue(`{"textures":{"SKIN":{"url":"http://textures.minecraft.net/texture/aa","metadata":{"model":"slim"}}}}`)},
		classic: {Type: storage.SkinTypeCustom, Identifier: "knight",
			Value: texturesValue(`{"textures":{"SKIN":{"url":"http://textures.minecraft.net/texture/bb"}}}`)},
		licensed: {Type: storage.SkinTypePlayer, Identifier: account.String()},
		fetched: {Type: storage.SkinTypePlayer, Identifier: account.String(),
			Value: texturesValue(`{"textures":{"SKIN":{"url":"http://textures.minecraft.net/texture/cc"}}}`)},
		broken: {Type: storage.SkinTypeURL, Identifier: "https://lania.example/b.png", Value: ptr("not base64!")},
	}

	service := NewSkinService(mapSkinsRestorerStorage(records), &outputConsole{})
	skins, err := service.GetPlayerSkins(context.Background(), uuid.UUIDs{slim, classic, licensed, fetched, broken, uuid.New()})
	if err != nil {
		t.Fatal(err)
	}

	if len(skins) != 4 {
		t.Fatalf("skins = %v, want the four readable ones", skins)
	}
	if skin := skins[slim]; skin.TextureURL != "https://textures.minecraft.net/texture/aa" || !skin.Slim || skin.MojangUUID != nil {
		t.Errorf("slim = %+v", skin)
	}
	if skin := skins[classic]; skin.TextureURL != "https://textures.minecraft.net/texture/bb" || skin.Slim {
		t.Errorf("classic = %+v", skin)
	}
	if skin := skins[licensed]; skin.TextureURL != "" || skin.MojangUUID == nil || *skin.MojangUUID != account {
		t.Errorf("licensed = %+v, want the account without a texture", skin)
	}
	if skin := skins[fetched]; skin.TextureURL != "https://textures.minecraft.net/texture/cc" || skin.MojangUUID == nil {
		t.Errorf("fetched = %+v", skin)
	}
}

type mapSkinsRestorerStorage map[uuid.UUID]*storage.SkinRecord

func (s mapSkinsRestorerStorage) FindPlayerSkins(_ context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*storage.SkinRecord, error) {
	records := make(map[uuid.UUID]*storage.SkinRecord)
	for _, mcUUID := range mcUUIDs {
		if record, ok := s[mcUUID]; ok {
			records[mcUUID] = record
		}
	}
	return records, nil
}

func TestSkinServiceSetPlayerSkin(t *testing.T) {
	mcUUID, account := uuid.New(), uuid.New()
	const link = "https://lania.example/v1/skins/files/0123456789abcdef.png"
	linkSkin := &storage.SkinRecord{Type: storage.SkinTypeURL, Identifier: link, Variant: ptr("SLIM"),
		Value: texturesValue(`{"textures":{"SKIN":{"url":"http://textures.minecraft.net/texture/aa","metadata":{"model":"slim"}}}}`)}
	accountSkin := &storage.SkinRecord{Type: storage.SkinTypePlayer, Identifier: account.String()}

	tests := []struct {
		name         string
		skin         string
		variant      string
		mojangUUID   *uuid.UUID
		states       []*storage.SkinRecord
		wantCommands []string
		wantErr      error
	}{
		{
			name: "link applies after a few reads", skin: link, variant: SkinVariantSlim,
			states:       []*storage.SkinRecord{nil, nil, linkSkin},
			wantCommands: []string{`skin set "` + link + `" ` + mcUUID.String() + ` slim`},
		},
		{
			name: "link the player already wears runs nothing", skin: link, variant: SkinVariantSlim,
			states: []*storage.SkinRecord{linkSkin},
		},
		{
			name: "other variant of the same link is applied", skin: link, variant: SkinVariantClassic,
			states:       []*storage.SkinRecord{linkSkin},
			wantCommands: []string{`skin set "` + link + `" ` + mcUUID.String() + ` classic`},
			wantErr:      ErrSkinNotApplied,
		},
		{
			name: "nickname waits for the account", skin: "Notch", variant: SkinVariantSlim, mojangUUID: &account,
			states:       []*storage.SkinRecord{linkSkin, accountSkin},
			wantCommands: []string{`skin set "Notch" ` + mcUUID.String()},
		},
		{
			name: "nickname without its account", skin: "Notch",
			states:  []*storage.SkinRecord{nil},
			wantErr: ErrInvalidSkin,
		},
		{
			name: "quote in a link", skin: `https://lania.example/a.png" ` + uuid.NewString(),
			states:  []*storage.SkinRecord{nil},
			wantErr: ErrInvalidSkin,
		},
		{
			name: "unknown variant", skin: link, variant: "wide",
			states:  []*storage.SkinRecord{nil},
			wantErr: ErrInvalidSkin,
		},
		{
			name: "skin SkinsRestorer never applies", skin: link,
			states:       []*storage.SkinRecord{nil},
			wantCommands: []string{`skin set "` + link + `" ` + mcUUID.String()},
			wantErr:      ErrSkinNotApplied,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &outputConsole{}
			service := newTestSkinService(&fakeSkinsRestorerStorage{mcUUID: mcUUID, states: tt.states}, console)

			skin, err := service.SetPlayerSkin(context.Background(), mcUUID, tt.skin, tt.variant, tt.mojangUUID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && skin == nil {
				t.Error("skin = nil, want the applied skin")
			}
			if len(console.commands) != len(tt.wantCommands) {
				t.Fatalf("commands = %q, want %q", console.commands, tt.wantCommands)
			}
			for i, command := range tt.wantCommands {
				if console.commands[i] != command {
					t.Errorf("command = %q, want %q", console.commands[i], command)
				}
			}
		})
	}
}

func TestSkinServiceClearPlayerSkin(t *testing.T) {
	mcUUID := uuid.New()
	worn := &storage.SkinRecord{Type: storage.SkinTypeCustom, Identifier: "knight"}

	tests := []struct {
		name         string
		states       []*storage.SkinRecord
		wantCommands int
		wantErr      error
	}{
		{name: "skin is taken off", states: []*storage.SkinRecord{worn, worn, nil}, wantCommands: 1},
		{name: "no skin runs nothing", states: []*storage.SkinRecord{nil}},
		{name: "skin stays", states: []*storage.SkinRecord{worn}, wantCommands: 1, wantErr: ErrSkinNotApplied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &outputConsole{}
			service := newTestSkinService(&fakeSkinsRestorerStorage{mcUUID: mcUUID, states: tt.states}, console)

			err := service.ClearPlayerSkin(context.Background(), mcUUID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if len(console.commands) != tt.wantCommands || (tt.wantCommands == 1 && console.commands[0] != "skin clear "+mcUUID.String()) {
				t.Errorf("commands = %q", console.commands)
			}
		})
	}
}
