package shadow

// FileType represents the type of a file system node.
type FileType string

const (
	FileTypeDir  FileType = "dir"
	FileTypeFile FileType = "file"
)

// TreeNode represents a file or directory in the docs tree.
// See SHADOW_VIEWER_specs.md Section 3.2
type TreeNode struct {
	Name     string     `json:"name"`
	Type     FileType   `json:"type"`
	Path     string     `json:"path,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

// TreeResponse is the response for the docs tree endpoint.
// GET /api/v1/shadow/docs/tree
type TreeResponse struct {
	Roots []TreeNode `json:"roots"`
}

// ContentResponse is the response for the docs content endpoint.
// GET /api/v1/shadow/docs/content
type ContentResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
