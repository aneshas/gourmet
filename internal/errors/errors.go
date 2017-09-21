// Package errors provides gourmet errors
package errors

import (
	"fmt"
)

// New creates new gourmet error
func New(status int, text, desc string) *Error {
	return &Error{status, text, desc}
}

// Error represents gourmet error type
type Error struct {
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Description string `json:"description"`
}

// Error implements error interface
func (e Error) Error() string {
	return fmt.Sprintf("%d %s <%s>", e.Status, e.StatusText, e.Description)
}
