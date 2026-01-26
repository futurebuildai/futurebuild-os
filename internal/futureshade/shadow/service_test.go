package shadow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetContent_PathTraversal tests that path traversal attacks are blocked.
// SECURITY: See SHADOW_VIEWER_specs.md Section 6.2 and 7.1
func TestGetContent_PathTraversal(t *testing.T) {
	// Create a temp directory structure for testing
	tempDir, err := os.MkdirTemp("", "shadow-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test directories and files
	docsDir := filepath.Join(tempDir, "docs")
	specsDir := filepath.Join(tempDir, "specs")
	secretDir := filepath.Join(tempDir, "secret")

	require.NoError(t, os.MkdirAll(docsDir, 0755))
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.MkdirAll(secretDir, 0755))

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "README.md"), []byte("# Docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "API.md"), []byte("# API Spec"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(secretDir, "secrets.txt"), []byte("SECRET DATA"), 0644))

	service := NewDocsService(tempDir)
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		wantErr error
		desc    string
	}{
		{
			name:    "direct traversal to root",
			path:    "../../../etc/passwd",
			wantErr: ErrPathTraversal,
			desc:    "Should block attempts to escape to /etc/passwd",
		},
		{
			name:    "embedded traversal in docs",
			path:    "docs/../../../etc/passwd",
			wantErr: ErrPathTraversal,
			desc:    "Should block traversal embedded after valid prefix",
		},
		{
			name:    "traversal to sibling secret dir",
			path:    "docs/../secret/secrets.txt",
			wantErr: ErrPathTraversal,
			desc:    "Should block traversal to sibling directories",
		},
		{
			name:    "invalid root - src directory",
			path:    "src/main.go",
			wantErr: ErrInvalidPath,
			desc:    "Should reject paths not starting with docs/ or specs/",
		},
		{
			name:    "invalid root - absolute path",
			path:    "/etc/passwd",
			wantErr: ErrInvalidPath,
			desc:    "Should reject absolute paths",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: ErrInvalidPath,
			desc:    "Should reject empty paths",
		},
		{
			name:    "directory not file",
			path:    "docs",
			wantErr: ErrInvalidPath,
			desc:    "Should reject paths that are directories",
		},
		{
			name:    "valid docs path",
			path:    "docs/README.md",
			wantErr: nil,
			desc:    "Should allow valid docs/ paths",
		},
		{
			name:    "valid specs path",
			path:    "specs/API.md",
			wantErr: nil,
			desc:    "Should allow valid specs/ paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := service.GetContent(ctx, tt.path)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, tt.desc)
				assert.Nil(t, content)
			} else {
				assert.NoError(t, err, tt.desc)
				assert.NotNil(t, content)
				assert.NotEmpty(t, content.Content)
			}
		})
	}
}

// TestGetContent_ValidPaths tests that valid paths work correctly.
func TestGetContent_ValidPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shadow-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	docsDir := filepath.Join(tempDir, "docs")
	subDir := filepath.Join(docsDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	testContent := "# Test Document\n\nThis is a test."
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "test.md"), []byte(testContent), 0644))

	service := NewDocsService(tempDir)
	ctx := context.Background()

	content, err := service.GetContent(ctx, "docs/subdir/test.md")
	require.NoError(t, err)

	assert.Equal(t, "docs/subdir/test.md", content.Path)
	assert.Equal(t, testContent, content.Content)
}

// TestGetTree_ReturnsAllowedRoots tests that GetTree only returns allowed directories.
func TestGetTree_ReturnsAllowedRoots(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shadow-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create directories
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "docs"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "specs"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "secret"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "src"), 0755))

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "docs", "README.md"), []byte("# Docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "specs", "API.md"), []byte("# API"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "secret", "secrets.txt"), []byte("SECRET"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "src", "main.go"), []byte("package main"), 0644))

	service := NewDocsService(tempDir)
	ctx := context.Background()

	tree, err := service.GetTree(ctx)
	require.NoError(t, err)

	// Should only have docs and specs roots
	assert.Len(t, tree.Roots, 2)

	rootNames := make(map[string]bool)
	for _, root := range tree.Roots {
		rootNames[root.Name] = true
	}

	assert.True(t, rootNames["docs"], "Should include docs/")
	assert.True(t, rootNames["specs"], "Should include specs/")
	assert.False(t, rootNames["secret"], "Should NOT include secret/")
	assert.False(t, rootNames["src"], "Should NOT include src/")
}

// TestGetTree_FiltersMdFiles tests that only .md files are included.
func TestGetTree_FiltersMdFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shadow-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Create various file types
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "README.md"), []byte("# Docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "config.json"), []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, ".hidden.md"), []byte("hidden"), 0644))

	service := NewDocsService(tempDir)
	ctx := context.Background()

	tree, err := service.GetTree(ctx)
	require.NoError(t, err)
	require.Len(t, tree.Roots, 1)

	docsRoot := tree.Roots[0]
	assert.Equal(t, "docs", docsRoot.Name)

	// Should only have README.md (not config.json or .hidden.md)
	assert.Len(t, docsRoot.Children, 1)
	assert.Equal(t, "README.md", docsRoot.Children[0].Name)
}
