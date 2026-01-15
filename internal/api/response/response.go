// Package response provides standardized API response helpers.
// See PRODUCTION_PLAN.md Task 2 (Standardize API Error Handling).
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorEnvelope is the standardized error response format.
// Format: {"error": {"code": 4xx, "message": "..."}}
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error code and message.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSONError writes a standardized JSON error response.
// ENGINEERING STANDARD: All API errors must use this helper for consistency.
func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    status,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("response: failed to encode error response", "error", err)
	}
}
