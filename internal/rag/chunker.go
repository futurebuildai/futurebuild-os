package rag

import (
	"strings"
)

// DefaultChunkSize is the target size for chunks in characters
const DefaultChunkSize = 1000

// DefaultOverlap is the overlap between chunks in characters
const DefaultOverlap = 100

// Chunker handles splitting text into semantic chunks
type Chunker struct {
	ChunkSize int
	Overlap   int
}

// NewChunker creates a new Chunker with default settings
func NewChunker() *Chunker {
	return &Chunker{
		ChunkSize: DefaultChunkSize,
		Overlap:   DefaultOverlap,
	}
}

// ChunkDocument splits a text into overlapping chunks
// This is a simple implementation that respects sentence boundaries where possible
// but primarily relies on character counts for now as a baseline.
func (c *Chunker) ChunkDocument(text string) []string {
	if text == "" {
		return []string{}
	}

	var chunks []string
	runes := []rune(text)
	totalLen := len(runes)

	if totalLen <= c.ChunkSize {
		return []string{text}
	}

	start := 0
	for start < totalLen {
		end := start + c.ChunkSize
		if end >= totalLen {
			end = totalLen
			chunks = append(chunks, string(runes[start:end]))
			break
		}

		// Try to find a sentence break or newline near the end to avoid cutting words
		// We look back up to 20% of chunk size
		lookbackLimit := int(float64(c.ChunkSize) * 0.2)
		foundBreak := false

		for i := 0; i < lookbackLimit; i++ {
			curr := end - i
			if curr <= start {
				break
			}
			r := runes[curr]
			// Check for sentence enders or newlines
			if r == '.' || r == '!' || r == '?' || r == '\n' {
				end = curr + 1 // Include the punctuation
				foundBreak = true
				break
			}
		}

		// If no natural break found, we just cut at ChunkSize (or maybe try space)
		if !foundBreak {
			// Try finding a space at least
			for i := 0; i < lookbackLimit; i++ {
				curr := end - i
				if curr <= start {
					break
				}
				if runes[curr] == ' ' {
					end = curr + 1
					foundBreak = true
					break
				}
			}
		}

		chunkStr := string(runes[start:end])
		chunks = append(chunks, strings.TrimSpace(chunkStr))

		// Move start forward, respecting overlap
		// We want the next chunk to start 'Overlap' characters before the current 'end'
		// BUT we should also align with the break we found if possible.
		// For simplicity in this v1, we just subtract overlap from the hard end
		// unless we are already at the end.

		start = end - c.Overlap
		if start < 0 {
			start = 0 // Should not happen
		}
		// Ensure we are making progress to avoid infinite loops
		if start <= (end - c.ChunkSize) {
			// If overlap effectively cancels out progress (unlikely with this logic), force move
			start = end
		}
	}

	return chunks
}
