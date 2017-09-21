// Package errors provides gourmet errors
package errors

import (
	"fmt"
)

// NewHTTP creates new gourmet error
func NewHTTP(status int, text, desc string) *HTTPError {
	return &HTTPError{status, text, desc}
}

// HTTPError represents gourmet http error type
type HTTPError struct {
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Description string `json:"description"`
}

// Error implements error interface
func (e HTTPError) Error() string {
	return fmt.Sprintf("%d %s <%s>", e.Status, e.StatusText, e.Description)
}
