package rag

import (
	"context"

	"github.com/colton/futurebuild/pkg/ai"
)

// Embedder handles interaction with the embedding model
type Embedder struct {
	client ai.Client
}

// NewEmbedder creates a new Embedder service
func NewEmbedder(client ai.Client) *Embedder {
	return &Embedder{
		client: client,
	}
}

// GenerateEmbeddings generates vector embeddings for a batch of texts
// Note: The VertexClient interface currently supports single string embedding.
// We will iterate here, but in a real prod scenario, we might want a batch API if available.
func (e *Embedder) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	var embeddings [][]float32

	for _, text := range texts {
		emb, err := e.client.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}
