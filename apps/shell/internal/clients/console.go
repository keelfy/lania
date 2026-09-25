package clients

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gorcon/rcon"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
)

// ErrConsoleDisabled means no RCON address is configured.
var ErrConsoleDisabled = errors.New("rcon is not configured")

const (
	consoleDialTimeout = 3 * time.Second
	consoleIOTimeout   = 5 * time.Second
)

// Console runs commands on the Minecraft server console.
type Console interface {
	// Execute returns ErrConsoleDisabled when RCON is not configured.
	Execute(ctx context.Context, command string) (string, error)
}

type rconConsole struct{}

func NewConsole() Console {
	return &rconConsole{}
}

// Execute opens a connection per command. Commands are rare and the server
// may restart at any time, so there is no connection to keep alive.
func (c *rconConsole) Execute(ctx context.Context, command string) (string, error) {
	address := config.GetRconAddress()
	if address == "" {
		return "", ErrConsoleDisabled
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	conn, err := rcon.Dial(
		address,
		config.GetRconPassword(),
		rcon.SetDialTimeout(consoleDialTimeout),
		rcon.SetDeadline(consoleIOTimeout),
	)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	output, err := conn.Execute(command)
	if err == nil {
		logger.Infof(ctx, "rcon %q: %s", command, strings.TrimSpace(output))
	}
	return output, err
}
