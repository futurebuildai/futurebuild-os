package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultGitHubBaseURL = "https://api.github.com"
	maxDiffSize          = 10 * 1024 // 10KB truncation limit
	httpTimeout          = 30 * time.Second
)

// GitHubService implements GitHubServicer using the GitHub REST API.
// See docs/AUTOMATED_PR_REVIEW_PRD.md
type GitHubService struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubService creates a new GitHub service with the provided PAT.
func NewGitHubService(token string) *GitHubService {
	return &GitHubService{
		token: token,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		baseURL: defaultGitHubBaseURL,
	}
}

// FetchPRDiff retrieves the diff for a pull request.
// Truncates large diffs to 10KB and includes a file list summary.
// See docs/AUTOMATED_PR_REVIEW_PRD.md Security: "Large diffs: Truncate to 10KB"
func (s *GitHubService) FetchPRDiff(ctx context.Context, owner, repo string, prNumber int) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", s.baseURL, owner, repo, prNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	// Request raw diff format
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch diff: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("github API error: %d %s", resp.StatusCode, string(body))
	}

	// Read with size limit to prevent memory issues
	diff, err := io.ReadAll(io.LimitReader(resp.Body, maxDiffSize*2))
	if err != nil {
		return "", fmt.Errorf("read diff: %w", err)
	}

	diffStr := string(diff)

	// Sanitize diff (remove --- delimiters to prevent prompt injection)
	diffStr = sanitizeDiff(diffStr)

	// Truncate if needed
	if len(diffStr) > maxDiffSize {
		// Get file list for summary
		files, err := s.fetchPRFiles(ctx, owner, repo, prNumber)
		if err != nil {
			files = []string{"(failed to fetch file list)"}
		}

		truncated := diffStr[:maxDiffSize]
		summary := fmt.Sprintf("\n\n[TRUNCATED - Full diff exceeds 10KB]\n\nFiles changed (%d):\n- %s",
			len(files), strings.Join(files, "\n- "))
		diffStr = truncated + summary
	}

	return diffStr, nil
}

// fetchPRFiles retrieves the list of files changed in a PR.
func (s *GitHubService) fetchPRFiles(ctx context.Context, owner, repo string, prNumber int) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", s.baseURL, owner, repo, prNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var files []struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.Filename
	}
	return result, nil
}

// PostPRComment posts a comment on a pull request.
// Uses the issues API endpoint since PR comments are issue comments.
func (s *GitHubService) PostPRComment(ctx context.Context, owner, repo string, prNumber int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", s.baseURL, owner, repo, prNumber)

	payload := map[string]string{"body": body}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal comment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("github API error: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// sanitizeDiff removes content that could be used for prompt injection.
// Specifically removes "---" delimiters that might confuse AI models.
func sanitizeDiff(diff string) string {
	// Replace standalone "---" lines (common in diffs) with safer alternative
	// This prevents potential prompt injection via crafted file content
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		// Only sanitize lines that are exactly "---" (separator lines)
		// Keep lines like "--- a/file.go" which are part of diff syntax
		if line == "---" {
			lines[i] = "[separator]"
		}
	}
	return strings.Join(lines, "\n")
}
