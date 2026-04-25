package utils

import (
	"context"
	"fmt"
	"strings"
)

func FormatNameColorsToMiniMessages(ctx context.Context, colors []string) string {
	if len(colors) == 0 {
		return ""
	} else if len(colors) == 1 {
		return fmt.Sprintf("<color:%s>", colors[0])
	}
	return fmt.Sprintf("<gradient:%s>", strings.Join(colors, ":"))
}
