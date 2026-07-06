package storage

import "errors"

// Storage error types
var (
	// ErrUnavailable indicates storage service is not available
	ErrUnavailable = errors.New("storage service is not available")

	// ErrNotImplemented indicates storage operation is not implemented
	ErrNotImplemented = errors.New("storage operation is not implemented")

	// ErrNotFound indicates file not found in storage
	ErrNotFound = errors.New("file not found in storage")

	// ErrInvalidPath indicates invalid storage path
	ErrInvalidPath = errors.New("invalid storage path")

	// ErrFileTooLarge indicates the object exceeds the caller-provided read limit.
	ErrFileTooLarge = errors.New("file exceeds maximum allowed size")
)
