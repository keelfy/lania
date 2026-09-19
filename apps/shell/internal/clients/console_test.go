package clients

import (
	"context"
	"errors"
	"testing"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
)

func TestConsoleExecute(t *testing.T) {
	var received []string
	server := rcontest.NewServer(
		rcontest.SetSettings(rcontest.Settings{Password: "secret"}),
		rcontest.SetCommandHandler(func(c *rcontest.Context) {
			received = append(received, c.Request().Body())
			_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "synced").WriteTo(c.Conn())
		}),
	)
	defer server.Close()

	t.Setenv("RCON_ADDRESS", server.Addr())
	t.Setenv("RCON_PASSWORD", "secret")

	output, err := NewConsole().Execute(context.Background(), "lp sync")
	if err != nil {
		t.Fatal(err)
	}
	if output != "synced" || len(received) != 1 || received[0] != "lp sync" {
		t.Errorf("unexpected exchange: output=%q received=%v", output, received)
	}

	t.Setenv("RCON_PASSWORD", "wrong")
	if _, err := NewConsole().Execute(context.Background(), "lp sync"); err == nil {
		t.Errorf("wrong password must fail")
	}
}

func TestConsoleDisabled(t *testing.T) {
	t.Setenv("RCON_ADDRESS", "")

	_, err := NewConsole().Execute(context.Background(), "lp sync")
	if !errors.Is(err, ErrConsoleDisabled) {
		t.Errorf("expected ErrConsoleDisabled, got %v", err)
	}
}
