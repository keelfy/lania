package commands

import (
	"testing"

	"github.com/lania-smp/backend/internal/domain"
)

func TestClaimChunksCommandValidate(t *testing.T) {
	manyChunks := make([]domain.ChunkPos, domain.MaxClaimBatch+1)
	for i := range manyChunks {
		manyChunks[i] = domain.ChunkPos{X: i}
	}

	tests := []struct {
		name      string
		dimension string
		chunks    []domain.ChunkPos
		wantErr   bool
	}{
		{name: "valid", dimension: "minecraft_overworld", chunks: []domain.ChunkPos{{X: -40, Z: -33}, {X: -39, Z: -33}}},
		{name: "no dimension", dimension: "", chunks: []domain.ChunkPos{{}}, wantErr: true},
		{name: "dimension with a path", dimension: "../overworld", chunks: []domain.ChunkPos{{}}, wantErr: true},
		{name: "no chunks", dimension: "minecraft_overworld", wantErr: true},
		{name: "too many chunks", dimension: "minecraft_overworld", chunks: manyChunks, wantErr: true},
		{name: "duplicate chunk", dimension: "minecraft_overworld", chunks: []domain.ChunkPos{{X: 1, Z: 2}, {X: 1, Z: 2}}, wantErr: true},
		{name: "outside the border", dimension: "minecraft_overworld", chunks: []domain.ChunkPos{{X: maxChunkIndex + 1}}, wantErr: true},
		{name: "on the border", dimension: "minecraft_overworld", chunks: []domain.ChunkPos{{X: -maxChunkIndex, Z: maxChunkIndex}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ClaimChunksCommand{Dimension: tt.dimension, Chunks: tt.chunks}
			if err := cmd.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
