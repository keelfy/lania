package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
)

type outputConsole struct {
	output   string
	err      error
	commands []string
}

func (c *outputConsole) Execute(_ context.Context, command string) (string, error) {
	c.commands = append(c.commands, command)
	return c.output, c.err
}

func TestWhitelistServiceCommands(t *testing.T) {
	console := &outputConsole{}
	service := NewWhitelistService(console)

	if err := service.AddPlayer(context.Background(), uuid.New(), "Steve"); err != nil {
		t.Fatal(err)
	}
	if err := service.RemovePlayer(context.Background(), uuid.New(), "Steve"); err != nil {
		t.Fatal(err)
	}

	want := []string{"whitelist add Steve", "whitelist remove Steve"}
	if len(console.commands) != 2 || console.commands[0] != want[0] || console.commands[1] != want[1] {
		t.Errorf("commands = %v, want %v", console.commands, want)
	}
}

func TestWhitelistServiceCustomCommands(t *testing.T) {
	t.Setenv("WHITELIST_ADD_COMMAND", "easywl add {uuid} {username}")
	t.Setenv("WHITELIST_REMOVE_COMMAND", "easywl remove {username}")
	mcUUID := uuid.New()
	console := &outputConsole{}
	service := NewWhitelistService(console)

	if err := service.AddPlayer(context.Background(), mcUUID, "Steve"); err != nil {
		t.Fatal(err)
	}
	if err := service.RemovePlayer(context.Background(), mcUUID, "Steve"); err != nil {
		t.Fatal(err)
	}

	want := []string{"easywl add " + mcUUID.String() + " Steve", "easywl remove Steve"}
	if len(console.commands) != 2 || console.commands[0] != want[0] || console.commands[1] != want[1] {
		t.Errorf("commands = %v, want %v", console.commands, want)
	}
}

func TestWhitelistServiceFailures(t *testing.T) {
	tests := []struct {
		name    string
		console *outputConsole
		wantErr error
	}{
		{"rcon disabled", &outputConsole{err: clients.ErrConsoleDisabled}, clients.ErrConsoleDisabled},
		{"unknown player", &outputConsole{output: "That player does not exist"}, ErrInvalidUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewWhitelistService(tt.console).AddPlayer(context.Background(), uuid.New(), "Steve")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AddPlayer() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
