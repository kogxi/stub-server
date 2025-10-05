package httpstub

import "errors"

var (
	// ErrStubNotFound is returned when a requested stub does not exist.
	ErrStubNotFound error = errors.New("stub not found")
	// ErrMethodNotAllowed is returned when the HTTP method used is not allowed for the requested stub.
	ErrMethodNotAllowed error = errors.New("method not allowed")
)
