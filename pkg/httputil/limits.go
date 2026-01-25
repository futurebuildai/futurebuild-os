package httputil

// MaxBodySize is the maximum allowed request body size (1MB).
// Used to prevent DoS attacks via memory exhaustion.
const MaxBodySize = 1 << 20 // 1048576 bytes
