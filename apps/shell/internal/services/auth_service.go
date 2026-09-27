package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
	"github.com/lania-smp/shell/internal/storage"
)

// ErrPlayerNotRegistered means the player never registered in game, so there is no password to replace.
var ErrPlayerNotRegistered = errors.New("player is not registered in game")

// ErrPlayerLicensed means the player logs in with a licensed account and has no password.
var ErrPlayerLicensed = errors.New("player logs in with a licensed account")

// ErrInvalidPasswordHash means the hash is not one NavAuth can verify.
var ErrInvalidPasswordHash = errors.New("password hash is not a bcrypt hash")

type AuthService interface {
	// SetPassword replaces the in-game password of a registered unlicensed player, turns off its two-factor
	// login and kicks the player, so whoever is in game under the nickname has to log in with the new password.
	SetPassword(ctx context.Context, mcUUID uuid.UUID, username, passwordBcrypt string) error
}

type authService struct {
	navAuthStorage storage.NavAuthStorage
	console        clients.Console
}

func NewAuthService(navAuthStorage storage.NavAuthStorage, console clients.Console) AuthService {
	return &authService{navAuthStorage: navAuthStorage, console: console}
}

// bcryptPrefixes are the bcrypt versions NavAuth verifies.
var bcryptPrefixes = []string{"$2a$", "$2b$", "$2y$"}

func (s *authService) SetPassword(ctx context.Context, mcUUID uuid.UUID, username, passwordBcrypt string) error {
	if !isBcryptHash(passwordBcrypt) {
		return ErrInvalidPasswordHash
	}
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("%w: %q", ErrInvalidUsername, username)
	}

	required, found, err := s.navAuthStorage.FindCredentialsRequired(ctx, mcUUID)
	if err != nil {
		return err
	}
	if !found {
		return ErrPlayerNotRegistered
	}
	if !required {
		return ErrPlayerLicensed
	}
	if err := s.navAuthStorage.SavePassword(ctx, mcUUID, passwordBcrypt); err != nil {
		return err
	}
	s.kick(ctx, username)
	return nil
}

// kick disconnects the player from the network. The password is already changed, so a failure is only logged:
// the player has to log in with the new password on the next join anyway. The command does nothing for a player
// who is offline.
func (s *authService) kick(ctx context.Context, username string) {
	command := strings.ReplaceAll(config.GetKickCommand(), "{username}", username)
	if _, err := s.console.Execute(ctx, command); err != nil && !errors.Is(err, clients.ErrConsoleDisabled) {
		logger.Errorf(ctx, "failed to kick %s after the password change: %v", username, err)
	}
}

func isBcryptHash(value string) bool {
	if len(value) != 60 {
		return false
	}
	for _, prefix := range bcryptPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
