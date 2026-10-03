package commands

import (
	"testing"

	"github.com/lania-smp/backend/internal/domain"
)

func allPrices(amount float64) []PrivilegePrice {
	prices := make([]PrivilegePrice, len(domain.AllowedCurrencies))
	for i, currency := range domain.AllowedCurrencies {
		prices[i] = PrivilegePrice{Currency: currency, Amount: amount}
	}
	return prices
}

func TestValidatePrivilegePermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		permission string
		wantErr    bool
	}{
		{"homes.commands.*", false},
		{"essentials.fly", false},
		{"plugin.cmd-name_1", false},
		{"", true},
		{"*", true},
		{"group.admin", true},
		{"weight.100", true},
		{"prefix.100.[A]", true},
		{"suffix.100.x", true},
		{"meta.key.value", true},
		{"displayname.x", true},
		{"Homes.Commands", true},
		{"homes commands", true},
		{"homes;op", true},
		{string(make([]byte, 201)), true},
	}
	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			t.Parallel()
			if err := ValidatePrivilegePermission(tt.permission); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrivilegePermission(%q) error = %v, wantErr %v", tt.permission, err, tt.wantErr)
			}
		})
	}
}

func TestSavePrivilegeCommand_Validate(t *testing.T) {
	t.Parallel()

	valid := SavePrivilegeCommand{Name: "Homes", Names: domain.CosmeticNames{"ru": "Дома"}, Permission: "homes.commands.*", Prices: allPrices(99)}

	tests := []struct {
		name    string
		mutate  func(*SavePrivilegeCommand)
		wantErr bool
	}{
		{"valid", func(*SavePrivilegeCommand) {}, false},
		{"two decimals", func(command *SavePrivilegeCommand) { command.Prices = allPrices(99.99) }, false},
		{"prices differ per currency", func(command *SavePrivilegeCommand) { command.Prices[0].Amount = 100; command.Prices[1].Amount = 1.5 }, false},
		{"blank name", func(command *SavePrivilegeCommand) { command.Name = "" }, true},
		{"unsupported locale", func(command *SavePrivilegeCommand) { command.Names = domain.CosmeticNames{"es": "Casas"} }, true},
		{"reserved permission", func(command *SavePrivilegeCommand) { command.Permission = "group.admin" }, true},
		{"blank permission", func(command *SavePrivilegeCommand) { command.Permission = "" }, true},
		{"no prices", func(command *SavePrivilegeCommand) { command.Prices = nil }, true},
		{"missing currency", func(command *SavePrivilegeCommand) { command.Prices = command.Prices[1:] }, true},
		{"duplicate currency", func(command *SavePrivilegeCommand) { command.Prices[1].Currency = command.Prices[0].Currency }, true},
		{"unknown currency", func(command *SavePrivilegeCommand) { command.Prices[0].Currency = "JPY" }, true},
		{"zero price", func(command *SavePrivilegeCommand) { command.Prices[0].Amount = 0 }, true},
		{"negative price", func(command *SavePrivilegeCommand) { command.Prices[0].Amount = -1 }, true},
		{"three decimals", func(command *SavePrivilegeCommand) { command.Prices[0].Amount = 1.005 }, true},
		{"price too large", func(command *SavePrivilegeCommand) { command.Prices[0].Amount = 100000000 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			command := valid
			command.Prices = allPrices(99)
			tt.mutate(&command)
			if err := command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
