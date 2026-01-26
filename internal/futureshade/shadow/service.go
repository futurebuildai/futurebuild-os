package shadow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Errors for ShadowDocs service.
var (
	ErrInvalidPath   = errors.New("invalid path")
	ErrPathTraversal = errors.New("path traversal detected")
	ErrFileNotFound  = errors.New("file not found")
)

// AllowedRoots are the only directories accessible via ShadowDocs.
// See SHADOW_VIEWER_specs.md Section 6.2
var AllowedRoots = []string{"docs", "specs"}

// DocsService provides access to project documentation files.
// SECURITY: Implements strict path validation to prevent directory traversal.
type DocsService struct {
	basePath string // Project root directory
}

// NewDocsService creates a new ShadowDocs service.
func NewDocsService(basePath string) *DocsService {
	return &DocsService{basePath: basePath}
}

// GetTree recursively builds the file tree for allowed directories.
func (s *DocsService) GetTree(ctx context.Context) (*TreeResponse, error) {
	roots := make([]TreeNode, 0, len(AllowedRoots))

	for _, root := range AllowedRoots {
		rootPath := filepath.Join(s.basePath, root)

		// Check if directory exists
		info, err := os.Stat(rootPath)
		if err != nil || !info.IsDir() {
			continue
		}

		node, err := s.buildTree(rootPath, root)
		if err != nil {
			continue
		}
		roots = append(roots, node)
	}

	return &TreeResponse{Roots: roots}, nil
}

// buildTree recursively builds a TreeNode from the file system.
func (s *DocsService) buildTree(fullPath, relativePath string) (TreeNode, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return TreeNode{}, err
	}

	node := TreeNode{
		Name: info.Name(),
		Type: FileTypeFile,
	}

	if info.IsDir() {
		node.Type = FileTypeDir

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return node, nil
		}

		children := make([]TreeNode, 0, len(entries))
		for _, entry := range entries {
			// Skip hidden files and non-markdown files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childRelPath := filepath.Join(relativePath, entry.Name())
			childFullPath := filepath.Join(fullPath, entry.Name())

			// Only include .md files or directories
			if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			child, err := s.buildTree(childFullPath, childRelPath)
			if err != nil {
				continue
			}
			children = append(children, child)
		}

		// Sort children: directories first, then files, alphabetically
		sort.Slice(children, func(i, j int) bool {
			if children[i].Type != children[j].Type {
				return children[i].Type == FileTypeDir
			}
			return children[i].Name < children[j].Name
		})

		node.Children = children
	} else {
		// For files, include the path
		node.Path = relativePath
	}

	return node, nil
}

// GetContent reads a file's content with strict path validation.
// SECURITY: Implements path traversal protection per SHADOW_VIEWER_specs.md Section 6.2
func (s *DocsService) GetContent(ctx context.Context, path string) (*ContentResponse, error) {
	// 1. Reject empty path
	if path == "" {
		return nil, ErrInvalidPath
	}

	// 2. SECURITY: Reject paths containing ".."
	if strings.Contains(path, "..") {
		return nil, ErrPathTraversal
	}

	// 3. Clean the path (removes redundant slashes, etc.)
	cleanPath := filepath.Clean(path)

	// 4. SECURITY: Validate path starts with allowed root
	validRoot := false
	for _, root := range AllowedRoots {
		if strings.HasPrefix(cleanPath, root+string(filepath.Separator)) || cleanPath == root {
			validRoot = true
			break
		}
	}
	if !validRoot {
		return nil, ErrInvalidPath
	}

	// 5. Build full path
	fullPath := filepath.Join(s.basePath, cleanPath)

	// 6. SECURITY: Resolve to absolute path and verify still within allowed roots
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, ErrInvalidPath
	}

	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, ErrInvalidPath
	}

	// Verify resolved path is still within base directory
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		return nil, ErrPathTraversal
	}

	// Additional check: verify we're still in an allowed root
	relPath, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return nil, ErrInvalidPath
	}

	validRootResolved := false
	for _, root := range AllowedRoots {
		if strings.HasPrefix(relPath, root+string(filepath.Separator)) || relPath == root {
			validRootResolved = true
			break
		}
	}
	if !validRootResolved {
		return nil, ErrPathTraversal
	}

	// 7. Verify it's a file, not a directory
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, ErrInvalidPath
	}

	// 8. Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return &ContentResponse{
		Path:    cleanPath,
		Content: string(content),
	}, nil
}
