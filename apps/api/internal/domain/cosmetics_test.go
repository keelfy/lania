package domain

import "testing"

func TestCosmeticNames_Localized(t *testing.T) {
	t.Parallel()

	names := CosmeticNames{"ru": "Лес", "en": ""}
	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{"set locale", "ru", "Лес"},
		{"blank locale falls back", "en", "Forest"},
		{"missing locale falls back", "de", "Forest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := names.Localized(tt.locale, "Forest"); got != tt.want {
				t.Errorf("Localized() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCosmeticNames_LocalizedNil(t *testing.T) {
	t.Parallel()

	var names CosmeticNames
	if got := names.Localized("ru", "Forest"); got != "Forest" {
		t.Errorf("Localized() = %q, want %q", got, "Forest")
	}
}
