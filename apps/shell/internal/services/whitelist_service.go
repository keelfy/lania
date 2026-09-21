package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/config"
)

type WhitelistService interface {
	// AddPlayer allows the player to join. It is idempotent.
	AddPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error
	// RemovePlayer forbids the player to join. It is idempotent.
	RemovePlayer(ctx context.Context, mcUUID uuid.UUID, username string) error
}

// ErrInvalidUsername means the username is not a valid Minecraft username.
var ErrInvalidUsername = errors.New("invalid minecraft username")

// usernamePattern is the Minecraft username format. It also keeps anything
// that could alter the command out of the console input.
var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)

// playerNotFoundOutput is what the server prints when it cannot resolve the username to a profile.
const playerNotFoundOutput = "That player does not exist"

type whitelistService struct {
	console clients.Console
}

func NewWhitelistService(console clients.Console) WhitelistService {
	return &whitelistService{console: console}
}

func (s *whitelistService) AddPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error {
	return s.run(ctx, config.GetWhitelistAddCommand(), mcUUID, username)
}

func (s *whitelistService) RemovePlayer(ctx context.Context, mcUUID uuid.UUID, username string) error {
	return s.run(ctx, config.GetWhitelistRemoveCommand(), mcUUID, username)
}

// run fails when the server cannot be reached. Unlike a permission sync, a
// missed whitelist change has no other source to recover from.
func (s *whitelistService) run(ctx context.Context, template string, mcUUID uuid.UUID, username string) error {
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("%w: %q", ErrInvalidUsername, username)
	}

	command := strings.NewReplacer("{username}", username, "{uuid}", mcUUID.String()).Replace(template)
	output, err := s.console.Execute(ctx, command)
	if err != nil {
		return fmt.Errorf("failed to run %q: %w", command, err)
	}
	if strings.Contains(output, playerNotFoundOutput) {
		return fmt.Errorf("%w: %q is unknown to the server", ErrInvalidUsername, username)
	}
	return nil
}
