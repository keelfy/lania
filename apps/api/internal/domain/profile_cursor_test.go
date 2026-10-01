package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestProfileCursorRoundTrip(t *testing.T) {
	value := "2025-01-02 03:04:05"
	for _, want := range []*ProfileCursor{
		{Value: &value, ID: uuid.New()},
		{ID: uuid.New()},
	} {
		got, err := DecodeProfileCursor(want.Encode())
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ID != want.ID || (got.Value == nil) != (want.Value == nil) || (want.Value != nil && *got.Value != *want.Value) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestDecodeProfileCursorRejectsGarbage(t *testing.T) {
	for _, token := range []string{"!!!", "bm90LWpzb24", "e30"} {
		if _, err := DecodeProfileCursor(token); err == nil {
			t.Errorf("token %q was accepted", token)
		}
	}
}
