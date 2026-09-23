package commands

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestSetProfileRoleCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command SetProfileRoleCommand
		wantErr bool
	}{
		{"staff role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: domain.RoleModerator}, false},
		{"player role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: domain.RolePlayer}, false},
		{"unknown role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: "vip"}, true},
		{"blank role", SetProfileRoleCommand{ProfileID: uuid.New()}, true},
		{"nil profile", SetProfileRoleCommand{Role: domain.RoleAdmin}, true},
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

func TestMergeProfilesCommand_Validate(t *testing.T) {
	t.Parallel()

	source, target := uuid.New(), uuid.New()
	tests := []struct {
		name    string
		command MergeProfilesCommand
		wantErr bool
	}{
		{"distinct profiles", MergeProfilesCommand{SourceProfileID: source, TargetProfileID: target}, false},
		{"same profile", MergeProfilesCommand{SourceProfileID: source, TargetProfileID: source}, true},
		{"nil source", MergeProfilesCommand{TargetProfileID: target}, true},
		{"nil target", MergeProfilesCommand{SourceProfileID: source}, true},
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

func TestSaveNameColorCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command SaveNameColorCommand
		wantErr bool
	}{
		{"valid gradient", SaveNameColorCommand{Name: "Forest", Colors: []string{"#16A34A", "#22C55E"}}, false},
		{"plain color", SaveNameColorCommand{Name: "Default"}, false},
		{"invalid color", SaveNameColorCommand{Name: "Forest", Colors: []string{"green"}}, true},
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

func TestSaveNamePrefixCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command SaveNamePrefixCommand
		wantErr bool
	}{
		{"s3 preview", SaveNamePrefixCommand{Name: "Popcat", Prefix: ":glyth_popcat:", Image: "s3://bucket/glyth_preview/popcat.png"}, false},
		{"https preview", SaveNamePrefixCommand{Name: "Popcat", Prefix: ":glyth_popcat:", Image: "https://cdn.example.com/popcat.png"}, false},
		{"no scheme", SaveNamePrefixCommand{Name: "Popcat", Prefix: ":glyth_popcat:", Image: "glyth_preview/popcat.png"}, true},
		{"unsupported scheme", SaveNamePrefixCommand{Name: "Popcat", Prefix: ":glyth_popcat:", Image: "ftp://bucket/popcat.png"}, true},
		{"blank image", SaveNamePrefixCommand{Name: "Popcat", Prefix: ":glyth_popcat:"}, true},
		{"blank prefix", SaveNamePrefixCommand{Name: "Popcat", Image: "s3://bucket/glyth_preview/popcat.png"}, true},
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

func TestSaveProductCommand_Validate(t *testing.T) {
	t.Parallel()

	cosmeticID := uuid.New()
	easyDonateID := int64(123)
	localizations := []SaveProductLocalization{
		{Locale: "ru", Name: "Лес", Description: "Зелёный градиент"},
		{Locale: "en", Name: "Forest", Description: "Green gradient"},
	}
	valid := SaveProductCommand{Category: domain.ProductCategoryNameColor, CosmeticID: &cosmeticID, PriceName: domain.ProductPriceNameNameColor, IsActive: true, EasyDonateProductID: &easyDonateID, Localizations: localizations}

	tests := []struct {
		name    string
		mutate  func(*SaveProductCommand)
		wantErr bool
	}{
		{"published product", func(*SaveProductCommand) {}, false},
		{"draft without EasyDonate ID", func(command *SaveProductCommand) { command.IsActive = false; command.EasyDonateProductID = nil }, false},
		{"published without EasyDonate ID", func(command *SaveProductCommand) { command.EasyDonateProductID = nil }, true},
		{"wrong tariff", func(command *SaveProductCommand) { command.PriceName = domain.ProductPriceNameNamePrefix }, true},
		{"missing cosmetic", func(command *SaveProductCommand) { command.CosmeticID = nil }, true},
		{"missing English localization", func(command *SaveProductCommand) { command.Localizations = command.Localizations[:1] }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			command := valid
			command.Localizations = append([]SaveProductLocalization(nil), valid.Localizations...)
			tt.mutate(&command)
			if err := command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateEDProductCommand_Validate(t *testing.T) {
	t.Parallel()

	valid := CreateEDProductCommand{
		SessionKey:  "session-key",
		CSRFToken:   "csrf-token",
		Name:        "Лес",
		Description: "Зелёный градиент",
		PriceName:   domain.ProductPriceNameNameColor,
	}

	tests := []struct {
		name    string
		mutate  func(*CreateEDProductCommand)
		wantErr bool
	}{
		{"valid without image", func(*CreateEDProductCommand) {}, false},
		{"valid with image", func(command *CreateEDProductCommand) { command.Image = []byte{1, 2, 3} }, false},
		{"missing session key", func(command *CreateEDProductCommand) { command.SessionKey = "" }, true},
		{"missing csrf token", func(command *CreateEDProductCommand) { command.CSRFToken = "" }, true},
		{"missing name", func(command *CreateEDProductCommand) { command.Name = "" }, true},
		{"missing description", func(command *CreateEDProductCommand) { command.Description = "" }, true},
		{"invalid price name", func(command *CreateEDProductCommand) { command.PriceName = "invalid" }, true},
		{"image too large", func(command *CreateEDProductCommand) { command.Image = make([]byte, MaxEDProductImageBytes+1) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			command := valid
			tt.mutate(&command)
			if err := command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
