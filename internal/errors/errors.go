// Package errors provides gourmet errors
package errors

import (
	"fmt"
)

// New creates new gourmet error
func New(status int, text, desc string) *Error {
	return &Error{status, text, desc}
}

// Error represents gourmet http error
type Error struct {
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Description string `json:"description"`
}

// Error returns error string
func (e Error) Error() string {
	return fmt.Sprintf("%d %s <%s>", e.Status, e.StatusText, e.Description)
}
